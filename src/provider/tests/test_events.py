"""测试事件总线（来自 quanttide_connect）。"""

from quanttide_connect.events import DomainEvent, EventBus


class _Handler:
    def __init__(self) -> None:
        self.events: list[DomainEvent] = []

    def handle(self, event: DomainEvent) -> None:
        self.events.append(event)


class _Event(DomainEvent):
    def __init__(self, name: str) -> None:
        self.name = name


class TestEventBus:
    def test_publish_to_protocol_handler(self) -> None:
        bus = EventBus()
        h = _Handler()
        bus.register(h)
        bus.publish(_Event("a"))
        assert len(h.events) == 1

    def test_publish_to_callable(self) -> None:
        bus = EventBus()
        received: list[DomainEvent] = []
        bus.register(received.append)
        bus.publish(_Event("b"))
        assert len(received) == 1

    def test_multiple_handlers(self) -> None:
        bus = EventBus()
        h1 = _Handler()
        h2 = _Handler()
        bus.register(h1)
        bus.register(h2)
        bus.publish(_Event("c"))
        assert len(h1.events) == 1
        assert len(h2.events) == 1
