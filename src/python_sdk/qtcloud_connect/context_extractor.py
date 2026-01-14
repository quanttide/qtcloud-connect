"""
Context 提取器
从 Memo 序列中提取认知语境
"""

from .memo import Memo
from .context import Context
from datetime import datetime


PROMPT_TEMPLATE = """{raw}

---

提取原始文本的认知语境，包括共同知识、经验图式、心理假设等。"""


class ContextExtractor:
    """Context 提取器"""

    def __init__(self):
        pass
    
    def extract(self, memo: Memo) -> Context:
        """
        从单个 Memo 提取 Context
        
        Args:
            memo: 单条 Memo
            
        Returns:
            提取的 Context（包含共同知识、经验图式、心理假设等维度）
        """
        # 构建完整提示词
        prompt = self._build_prompt(memo.content)
        
        # 调用 LLM 提取（待接入）
        # llm_response = self._call_llm(prompt)
        # context = self._parse_llm_response(llm_response)
        # return context
        
        # 临时返回模拟结果
        return Context(
            id=self._generate_context_id(),
            summary=f"基于 memo {memo.id} 的认知语境",
            shared_knowledge="共同知识：利他是高级利己",
            mental_schemas="经验图式：价值观驱动行动",
            psychological_assumptions="心理假设：寻求价值闭环",
            timestamp=datetime.now()
        )
    
    def _build_prompt(self, content: str) -> str:
        """构建完整提示词"""
        # 替换模板中的 {raw} 占位符
        return PROMPT_TEMPLATE.replace("{raw}", content)
    
    def _generate_context_id(self) -> str:
        """生成 context ID"""
        return f"ctx_{datetime.now().strftime('%Y%m%d%H%M%S')}"
