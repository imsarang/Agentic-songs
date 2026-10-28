from typing import Literal

from pydantic import BaseModel, Field, HttpUrl

Period = Literal["Q1", "Q2", "Q3", "Q4", "FY"]
Scope = Literal["standalone", "consolidated"]
DocType=Literal["annual_report", "results", "presentation", "transcript"]
Kind=Literal["narrative", "statement"]
StatementName=Literal["balance_sheet", "profit_and_loss", "cash_flow"]

class Source(BaseModel):
    id: str = Field(pattern = r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
    url: HttpUrl
    symbol: str = Field(pattern=r"^[A-Z0-9]+$")
    fiscal_year: int = Field(ge=1990, le=2100)
    period: Period
    scope: Scope
    doc_type: DocType
    kind: Kind

class Fetched(BaseModel):
    body: bytes
    content_type: str

class StatementRow(BaseModel):
    statement: StatementName
    line: str = Field(min_length=1)
    value: float
    unit: str = "INR"
    segment: str = ""

class Derived(BaseModel):
    kind: Kind
    body: bytes
    content_type: str
    suffix: str

class QueueMessage(BaseModel):
    source_id: str
    sha256: str
    key: str
    kind: Kind
    symbol: str
    fiscal_year: int
    period: Period
    scope: Scope
    doc_type: DocType
    url: str

class CheckError(Exception):
    pass
