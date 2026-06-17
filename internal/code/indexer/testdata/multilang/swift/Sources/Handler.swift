import Foundation

class Handler {
    func health() -> String {
        return "ok"
    }
}

protocol Store {
    func save(key: String)
}

func bootstrap() -> String {
    return "ready"
}
