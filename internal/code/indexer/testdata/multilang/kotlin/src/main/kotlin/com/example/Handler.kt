package com.example

import java.util.List

class Handler {
    fun health(): String = "ok"

    fun names(): List<String> = listOf("a")
}

interface Store {
    fun save(key: String)
}
