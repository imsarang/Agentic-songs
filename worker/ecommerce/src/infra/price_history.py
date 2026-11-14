from decimal import Decimal

from infra.db import prices_db
from infra.ports import Prices


class PriceHistory(Prices):
    def __init__(self) -> None:
        self._db = prices_db

    def record(self, host: str, external_id: str, price: Decimal | str, currency: str | None = "INR") -> bool:
        amount = Decimal(price)
        code = currency or "INR"
        latest = self._db.run(
            """
            select price from price_history
            where host = %s and external_id = %s
            order by observed_at desc
            limit 1
            """,
            (host, external_id),
            one=True,
        )
        if latest is not None and latest["price"] == amount:
            return False
        self._db.run(
            """
            insert into price_history (host, external_id, price, currency)
            values (%s, %s, %s, %s)
            """,
            (host, external_id, amount, code),
        )
        return True