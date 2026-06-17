#include <string>

class Handler {
public:
    std::string health() {
        return "ok";
    }
};

std::string bootstrap() {
    return "ready";
}
