[CmdletBinding()]
param(
    [int]$Port = 18080
)

$ErrorActionPreference = 'Stop'

$AppRoot = Split-Path -Parent $PSScriptRoot
$ProviderDir = Join-Path $AppRoot 'src/provider'
$CliDir = Join-Path $AppRoot 'src/cli'
$StudioDir = Join-Path $AppRoot 'src/studio'
$Endpoint = "http://127.0.0.1:$Port/api"
$RunId = [Guid]::NewGuid().ToString('N')
$TempRoot = [System.IO.Path]::GetTempPath()
$DbPath = Join-Path $TempRoot "qtcloud-connect-v0.1-$RunId.db"
$ProviderBinaryName = "qtcloud-connect-provider-v0.1-$RunId"
if ($IsWindows) {
    $ProviderBinaryName += '.exe'
}
$ProviderBinary = Join-Path $TempRoot $ProviderBinaryName
$ProviderProcess = $null

function Assert-Condition {
    param(
        [bool]$Condition,
        [string]$Message
    )

    if (-not $Condition) {
        throw "v0.1 verification failed: $Message"
    }
}

function Invoke-JsonRequest {
    param(
        [ValidateSet('Get', 'Post', 'Put')]
        [string]$Method,
        [string]$Uri,
        [hashtable]$Body
    )

    $request = @{
        Method = $Method
        Uri = $Uri
        TimeoutSec = 5
    }
    if ($null -ne $Body) {
        $request.ContentType = 'application/json'
        $request.Body = $Body | ConvertTo-Json -Compress
    }
    Invoke-RestMethod @request
}

function Invoke-CliJson {
    param(
        [string[]]$Arguments
    )

    Push-Location $CliDir
    try {
        $output = & cargo run --quiet --locked -- @Arguments
        Assert-Condition ($LASTEXITCODE -eq 0) "CLI command failed: $($Arguments -join ' ')"
        $json = $output -join [Environment]::NewLine
        Assert-Condition (-not [string]::IsNullOrWhiteSpace($json)) 'CLI returned no JSON output'
        $json | ConvertFrom-Json
    } finally {
        Pop-Location
    }
}

function Invoke-ProviderBuild {
    Push-Location $ProviderDir
    try {
        & go build -o $ProviderBinary ./cmd/server
        Assert-Condition ($LASTEXITCODE -eq 0) 'Provider build failed'
    } finally {
        Pop-Location
    }
}

function Invoke-StudioBuild {
    Push-Location $StudioDir
    try {
        & flutter pub get
        Assert-Condition ($LASTEXITCODE -eq 0) 'Studio dependencies failed'
        & flutter build web --release --base-href /
        Assert-Condition ($LASTEXITCODE -eq 0) 'Studio web build failed'
    } finally {
        Pop-Location
    }
}

try {
    Assert-Condition ($null -ne (Get-Command go -ErrorAction SilentlyContinue)) 'go is not installed'
    Assert-Condition ($null -ne (Get-Command cargo -ErrorAction SilentlyContinue)) 'cargo is not installed'
    Assert-Condition ($null -ne (Get-Command flutter -ErrorAction SilentlyContinue)) 'flutter is not installed'

    Invoke-ProviderBuild
    Invoke-StudioBuild

    $oldDbPath = $env:DB_PATH
    $oldPort = $env:PORT
    $env:DB_PATH = $DbPath
    $env:PORT = $Port.ToString()
    try {
        $ProviderProcess = Start-Process `
            -FilePath $ProviderBinary `
            -WorkingDirectory $ProviderDir `
            -PassThru `
            -WindowStyle Hidden

        $healthy = $false
        for ($attempt = 0; $attempt -lt 30; $attempt++) {
            try {
                $health = Invoke-RestMethod `
                    -Method Get `
                    -Uri "http://127.0.0.1:$Port/healthz" `
                    -TimeoutSec 2
                if ($health.status -eq 'ok') {
                    $healthy = $true
                    break
                }
            } catch {
                Start-Sleep -Seconds 1
            }
        }
        Assert-Condition $healthy 'Provider did not become healthy'

    $created = Invoke-CliJson @(
        'consensus',
        'create',
        '--endpoint',
            $Endpoint,
            '--title',
            'v0.1 验收共识',
            '--description',
            '验证 CLI 写入、Provider 持久化和 Studio 展示闭环。'
    )
    Assert-Condition (-not [string]::IsNullOrWhiteSpace($created.id)) 'created consensus has no id'
    Assert-Condition ($created.status -eq 'proposed') 'created consensus is not proposed'
    Assert-Condition ($created.description -eq '验证 CLI 写入、Provider 持久化和 Studio 展示闭环。') 'created consensus description is incorrect'

    $shown = Invoke-CliJson @(
        'consensus',
        'show',
        '--endpoint',
        $Endpoint,
        $created.id
    )
    Assert-Condition ($shown.id -eq $created.id) 'CLI show returned the wrong consensus'

    $updated = Invoke-CliJson @(
        'consensus',
        'update',
        '--endpoint',
        $Endpoint,
        '--title',
        'v0.1 验收共识（更新）',
        $created.id
    )
    Assert-Condition ($updated.title -eq 'v0.1 验收共识（更新）') 'CLI update did not change the title'
    Assert-Condition ($updated.description -eq '验证 CLI 写入、Provider 持久化和 Studio 展示闭环。') 'CLI update cleared the existing description'

    $secondCreated = Invoke-CliJson @(
        'consensus',
        'create',
        '--endpoint',
        $Endpoint,
        '--title',
        'v0.1 第二条验收共识',
        '--description',
        '用于验证图谱关系和废弃状态。'
    )
    Assert-Condition ($secondCreated.status -eq 'proposed') 'second consensus is not proposed'

    $listed = Invoke-CliJson @(
        'consensus',
            'list',
            '--endpoint',
            $Endpoint
        )
        Assert-Condition ($listed.total -ge 1) 'consensus list is empty after create'

        $confirmed = Invoke-CliJson @(
            'consensus',
            'confirm',
            '--endpoint',
            $Endpoint,
            $created.id
        )
        Assert-Condition ($confirmed.status -eq 'confirmed') 'consensus confirmation failed'

        $graph = Invoke-JsonRequest `
            -Method Post `
            -Uri "$Endpoint/consensus-graphs" `
            -Body @{
                name = 'v0.1 验收图'
                description = '验证共识图持久化。'
            }
        Assert-Condition (-not [string]::IsNullOrWhiteSpace($graph.id)) 'created graph has no id'

    $graphWithNode = Invoke-JsonRequest `
            -Method Post `
            -Uri "$Endpoint/consensus-graphs/$($graph.id)/nodes" `
            -Body @{ consensus_id = $created.id }
    Assert-Condition ($graphWithNode.nodes.id -contains $created.id) 'created consensus is missing from graph'

    $graphWithSecondNode = Invoke-JsonRequest `
        -Method Post `
        -Uri "$Endpoint/consensus-graphs/$($graph.id)/nodes" `
        -Body @{ consensus_id = $secondCreated.id }
    Assert-Condition (@($graphWithSecondNode.nodes).Count -eq 2) 'second consensus is missing from graph'

    $graphWithRelation = Invoke-JsonRequest `
        -Method Post `
        -Uri "$Endpoint/consensus-graphs/$($graph.id)/relations" `
        -Body @{
            from = $created.id
            to = $secondCreated.id
            relation_type = '验证'
        }
    Assert-Condition (@($graphWithRelation.edges).Count -eq 1) 'graph relation was not persisted'

    $positionedGraph = Invoke-JsonRequest `
        -Method Put `
        -Uri "$Endpoint/consensus-graphs/$($graph.id)/nodes/$($created.id)/position" `
        -Body @{ x = 318.5; y = 204.25 }
    Assert-Condition ($positionedGraph.node_positions.$($created.id).x -eq 318.5) 'graph node x position was not persisted'
    Assert-Condition ($positionedGraph.node_positions.$($created.id).y -eq 204.25) 'graph node y position was not persisted'

    $deprecated = Invoke-CliJson @(
        'consensus',
        'deprecate',
        '--endpoint',
        $Endpoint,
        $secondCreated.id
    )
    Assert-Condition ($deprecated.status -eq 'deprecated') 'CLI deprecate failed'

    $reloadedGraph = Invoke-JsonRequest `
        -Method Get `
        -Uri "$Endpoint/consensus-graphs/$($graph.id)"
    Assert-Condition (@($reloadedGraph.nodes).Count -eq 2) 'reloaded graph lost nodes'
    Assert-Condition (@($reloadedGraph.edges).Count -eq 1) 'reloaded graph lost relation'

    Write-Output 'v0.1 verification passed: Provider -> CLI -> Studio build and data contract'
    } finally {
        if ($null -ne $oldDbPath) {
            $env:DB_PATH = $oldDbPath
        } else {
            Remove-Item Env:DB_PATH -ErrorAction SilentlyContinue
        }
        if ($null -ne $oldPort) {
            $env:PORT = $oldPort
        } else {
            Remove-Item Env:PORT -ErrorAction SilentlyContinue
        }
    }
} finally {
    if ($null -ne $ProviderProcess) {
        Stop-Process -Id $ProviderProcess.Id -Force -ErrorAction SilentlyContinue
        $ProviderProcess.WaitForExit()
    }
    Remove-Item -LiteralPath $ProviderBinary -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $DbPath -Force -ErrorAction SilentlyContinue
}
