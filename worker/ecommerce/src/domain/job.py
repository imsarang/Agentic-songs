from dataclasses import dataclass


@dataclass
class Job:
    type: str
    origin: str
    category: str
    host: str
    name: str | None = None
    url: str | None = None
    key: str | None = None
    external_id: str | None = None
    handle: str | None = None
    title: str | None = None
    vendor: str | None = None
    description: str | None = None
    specs: dict | None = None
    price: str | int | float | None = None
    currency: str | None = None

    @classmethod
    def from_dict(cls, message: dict) -> "Job":
        return cls(
            type=message["type"],
            origin=message["origin"],
            category=message["category"],
            host=message["host"],
            name=message.get("name"),
            url=message.get("url"),
            key=message.get("key"),
            external_id=message.get("external_id"),
            handle=message.get("handle"),
            title=message.get("title"),
            vendor=message.get("vendor"),
            description=message.get("description"),
            specs=message.get("specs"),
            price=message.get("price"),
            currency=message.get("currency"),
        )
