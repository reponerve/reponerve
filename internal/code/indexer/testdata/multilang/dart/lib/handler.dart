import 'dart:convert';

class Handler {
  String health() => 'ok';

  void run() {
    health();
  }
}

String bootstrap() => 'ready';
