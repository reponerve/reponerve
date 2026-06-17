<?php

namespace App;

use Psr\Http\Message\ResponseInterface;

class Handler {
    public function health(): string {
        return 'ok';
    }
}

function bootstrap(): string {
    return 'ready';
}
