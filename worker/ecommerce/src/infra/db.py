import os
import psycopg
from psycopg.rows import dict_row

class Database:
    def __init__(self, dsn: str) -> None:
        self._dsn = dsn
        self._connection: psycopg.Connection | None = None

    def connect(self) -> psycopg.Connection:
        connection = self._connection
        if connection is None or connection.closed:
            self._connection = psycopg.connect(self._dsn, row_factory=dict_row)
        return self._connection

    def close(self) -> None:
        if self._connection is not None and not self._connection.closed:
            self._connection.close()
        self._connection = None

    def run(self, sql: str, params: tuple=(), *, one: bool = False):
        last_error: Exception | None
        
        # retry twice
        for _ in range(2):
            try:
                connection = self.connect()
                with connection.cursor() as cursor:
                    cursor.execute(sql, params)
                    if cursor.description is None:
                        row = None
                    elif one:
                        row = cursor.fetchone()
                    else:
                        row = cursor.fetchall()
                connection.commit()
                return row
            except (psycopg.OperationalError, psycopg.InterfaceError) as exc:
                self.close()
                last_error = exc
        raise last_error

catalog_db = Database(os.environ.get(
    "CATALOG_DSN",
    "postgresql://shanks:value@127.0.0.1:5433/catalog",
))
prices_db = Database(os.environ.get(
    "PRICES_DSN",
    "postgresql://shanks:value@127.0.0.1:5434/prices",
))