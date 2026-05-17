"""测试智能体业务逻辑（mock LLM API）。"""

from unittest.mock import MagicMock, patch

from quanttide_connect.models import ConsensusStatus, Role
from quanttide_connect.services.consensus import ConsensusService
from quanttide_connect.services.relation import RelationService

from app.agents.consensus_agent import ConsensusAgent
from app.agents.message_agent import MessageAgent
from app.storage import Storage


class TestMessageAgent:
    def setup_method(self) -> None:
        self.agent = MessageAgent()
        self.agent.api_key = "test-key"

    def test_consensus_summary_empty(self) -> None:
        summary = self.agent.get_consensus_summary([])
        assert "暂无已确认的共识" in summary

    def test_consensus_summary_with_items(self) -> None:
        consensuses = [{"content": "用 PostgreSQL"}, {"content": "Python 后端"}]
        summary = self.agent.get_consensus_summary(consensuses)
        assert "用 PostgreSQL" in summary

    @patch("app.agents.message_agent.requests.post")
    def test_reply_sends_correct_prompt(self, mock_post: MagicMock) -> None:
        mock_post.return_value.ok = True
        mock_post.return_value.json.return_value = {
            "choices": [{"message": {"content": "好的，就用 PostgreSQL。"}}]
        }
        reply = self.agent.reply(
            user_message="用什么数据库？",
            history=[],
            confirmed_consensuses=[{"content": "用 PostgreSQL"}],
        )
        sent = mock_post.call_args[1]["json"]
        assert "用 PostgreSQL" in sent["messages"][0]["content"]
        assert reply == "好的，就用 PostgreSQL。"


class TestConsensusAgent:
    def test_parse_action_full(self) -> None:
        agent = ConsensusAgent.__new__(ConsensusAgent)
        result = agent._parse_action(
            'action: propose\ncontent: PostgreSQL\nrelated_messages: ["m1"]\n'
        )
        assert result["action"] == "propose"
        assert result["content"] == "PostgreSQL"

    def test_parse_action_missing_action(self) -> None:
        agent = ConsensusAgent.__new__(ConsensusAgent)
        assert agent._parse_action("content: 测试") is None

    def test_execute_propose(self, storage: Storage) -> None:
        con_svc = ConsensusService(storage)
        rel_svc = RelationService(storage)
        agent = ConsensusAgent(storage, con_svc, rel_svc)
        agent.api_key = "test-key"
        agent._execute(
            {"action": "propose", "content": "PostgreSQL", "related_messages": []}
        )
        assert len(storage.list_consensuses(ConsensusStatus.proposed)) == 1

    def test_execute_confirm(self, storage: Storage) -> None:
        con_svc = ConsensusService(storage)
        rel_svc = RelationService(storage)
        agent = ConsensusAgent(storage, con_svc, rel_svc)
        agent.api_key = "test-key"
        c = con_svc.propose("PostgreSQL")
        agent._execute({"action": "confirm", "content": "PostgreSQL"})
        assert storage.get_consensus(c.id).status == ConsensusStatus.confirmed

    def test_execute_deprecate(self, storage: Storage) -> None:
        con_svc = ConsensusService(storage)
        rel_svc = RelationService(storage)
        agent = ConsensusAgent(storage, con_svc, rel_svc)
        agent.api_key = "test-key"
        c = con_svc.propose("PostgreSQL")
        con_svc.confirm(c.id)
        agent._execute({"action": "deprecate", "content": "PostgreSQL"})
        assert storage.get_consensus(c.id).status == ConsensusStatus.deprecated

    @patch("app.agents.consensus_agent.requests.post")
    def test_observe_propose(self, mock_post: MagicMock, storage: Storage) -> None:
        mock_post.return_value.ok = True
        mock_post.return_value.json.return_value = {
            "choices": [
                {
                    "message": {
                        "content": """
[CONSENSUS_ACTION]
action: propose
content: 团队用 PostgreSQL
related_messages: []
[/CONSENSUS_ACTION]
"""
                    }
                }
            ]
        }
        con_svc = ConsensusService(storage)
        rel_svc = RelationService(storage)
        agent = ConsensusAgent(storage, con_svc, rel_svc)
        agent.api_key = "test-key"
        from quanttide_connect.models import Message

        user_msg = Message(content="我们用 PostgreSQL", role=Role.user)
        agent_msg = Message(content="好的", role=Role.agent)
        agent.observe(user_msg, agent_msg, [])
        assert len(storage.list_consensuses(ConsensusStatus.proposed)) >= 1

    @patch("app.agents.consensus_agent.requests.post")
    def test_observe_no_action(self, mock_post: MagicMock, storage: Storage) -> None:
        mock_post.return_value.ok = True
        mock_post.return_value.json.return_value = {
            "choices": [{"message": {"content": "[NO_ACTION]"}}]
        }
        con_svc = ConsensusService(storage)
        rel_svc = RelationService(storage)
        agent = ConsensusAgent(storage, con_svc, rel_svc)
        agent.api_key = "test-key"
        from quanttide_connect.models import Message

        agent.observe(
            Message(content="hi", role=Role.user),
            Message(content="ok", role=Role.agent),
            [],
        )
        assert len(storage.list_consensuses()) == 0
