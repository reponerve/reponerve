package com.example.api;

import java.util.List;

public class Handler {
    public String health() {
        return "ok";
    }

    public List<String> names() {
        return List.of("a");
    }
}
