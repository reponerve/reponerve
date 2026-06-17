package com.example

import scala.collection.mutable

class Handler {
  def health(): String = "ok"
}

trait Store {
  def save(key: String): Unit
}

object Bootstrap {
  def run(): String = "ready"
}
