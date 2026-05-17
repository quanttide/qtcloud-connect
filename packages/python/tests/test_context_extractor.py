"""
Python SDK 测试
"""
from qtcloud_connect import ContextExtractor, Memo
from datetime import datetime


def test_context_extractor():
    """测试 Context 提取"""

    # 准备测试数据 - 使用提示词文档中的示例输入
    raw_text = """我觉得这里面有一个本质的问题，就是嗯我们一直就我们一直认为利他就纯粹是利他，但是其实利他是一种高级的利己。那么个人来说的话，我们愿意去帮助别人是因为我们可以从帮助别人的过程之中嗯成就自我的价值或者启发，这些是模糊的。但是企业是一个理性的动物，那我们需要什么样的理由才能够让企业能够直接的获得，因为利他还获得了正外部性的反馈，从而让我们更愿意去做这件事情

但我觉得你说的这些其实都没有回答根本的问题，就是其嗯比如说我们给高校老师服务，那么高校老师服务的产出到底对我们有什么价值？以至于我们要长期为他们的利益考虑呢

嗯我觉得我问了一个非常根本的问题，这个问题不是AI能回答的，我觉得你的回答都非常的粗糙


我觉得我们首先要去讨论我们为什么要去问这个问题，就是我们问这个问题的本质是因为呃团队的价值观是这样子的。如果我们不找到一套跟团队的价值观配套的方法论，那么团队没有办法从心里接受。嗯我我我想我想要在动力上把这种嗯嗯我们的产品推销给其他人

但是你这种做法又把它推到了道德道德高地，它不是我们所说的利他是最高级的利己这种方法论

对我觉得我在思考一个很本质的问题，就是嗯我们怎么在当下的行动之中找到一个新的切入口？这个新的切入口既能够满足我们的需求，又能够呃指引我们的方向

就我觉得强行去追求这个也也是不现实的，这不是一种科学的方法"""

    # 创建 Memo
    memo = Memo(
        id="1",
        content=raw_text,
        timestamp=datetime.now()
    )

    # 创建提取器
    extractor = ContextExtractor()

    # 执行提取
    context = extractor.extract(memo)

    # 验证结果
    assert context is not None
    assert context.summary is not None
    assert context.shared_knowledge is not None
    assert context.mental_schemas is not None
    assert context.psychological_assumptions is not None

    print(f"✅ Context 提取成功")
    print(f"   ID: {context.id}")
    print(f"   摘要: {context.summary}")
    print(f"   共同知识: {context.shared_knowledge}")
    print(f"   经验图式: {context.mental_schemas}")
    print(f"   心理假设: {context.psychological_assumptions}")


if __name__ == "__main__":
    test_context_extractor()
