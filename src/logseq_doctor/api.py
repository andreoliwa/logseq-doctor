from dataclasses import dataclass
from typing import List

import requests


@dataclass
class QueryResult:
    block_id: str
    journal_iso_date: int
    name: str
    url: str
    content: str


@dataclass
class LogseqApi:
    url: str
    token: str

    def query(self, query: str) -> List[QueryResult]:
        """Query Logseq API."""
        session = requests.Session()
        session.headers.update(
            {
                "Authorization": f"Bearer {self.token}",
                "Content-Type": "application/json",
            }
        )
        resp = session.post(f"{self.url}/api", json={"method": "logseq.db.q", "args": [query]})
        resp.raise_for_status()

        block_url = "logseq://graph/captains-log?block-id="
        rows: List[QueryResult] = []
        for obj in resp.json():
            page = obj.get("page", {})
            block_id = obj.get("uuid")
            rows.append(
                QueryResult(
                    block_id=block_id,
                    journal_iso_date=page.get("journalDay", 0),
                    name=page.get("originalName"),
                    url=f"{block_url}{block_id}",
                    content=obj.get("content").splitlines()[0],
                )
            )
        return rows
