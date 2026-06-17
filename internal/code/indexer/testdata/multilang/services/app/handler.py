from requests import get


def health_check() -> str:
    return "ok"


class Handler:
    def handle(self, path: str) -> str:
        return get(path).text
