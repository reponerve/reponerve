require "json"

class Handler
  def health
    "ok"
  end

  def handle(path)
    path
  end
end

def bootstrap
  Handler.new.health
end
