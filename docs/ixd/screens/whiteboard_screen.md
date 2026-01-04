# 白板页面

## UI设计

### 组件示意图

```
+------------------+------------------------+
|   NotePool       |      Workbench         |
|                  |                        |
| • Note1          |  +------------------+  |
| • Note2          |  |  AssemblyArea    |  |
| • Note3          |  |                  |  |
|                  |  +------------------+  |
|                  |                        |
|                  |  +------------------+  |
|                  |  |   MemoDraft      |  |
|                  |  |                  |  |
|                  |  +------------------+  |
+------------------+------------------------+
```

### 组件层级树

```
WhiteboardScreen
├── NotePool          # 便签池（原材料）
└── Workbench         # 工作台（操作台）
    ├── AssemblyArea  # 组装区（用户拖放、排列便签）
    └── MemoDraft     # 备忘录草稿（AI 生成 + 可编辑）
```

### Flutter代码示例

```
// whiteboard_screen.dart (简化版)
@override
Widget build(BuildContext context) {
  return Row(
    children: [
      // NotePool
      SizedBox(
        width: 320,
        child: ListView(
          children: [
            NoteCard(title: "要点1", content: "..."),
            NoteCard(title: "要点2", content: "..."),
          ],
        ),
      ),
      
      // Workbench
      Expanded(
        child: Column(
          children: [
            // AssemblyArea
            Expanded(
              flex: 4,
              child: Container(
                decoration: BoxDecoration(border: Border.all()),
                child: Center(child: Text("拖入便签")),
              ),
            ),
            
            // MemoDraft
            Expanded(
              flex: 6,
              child: Card(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Text("备忘录草稿将在这里生成..."),
                ),
              ),
            ),
          ],
        ),
      ),
    ],
  );
}
```

## UX设计
