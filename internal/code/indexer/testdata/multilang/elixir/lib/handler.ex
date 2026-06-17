defmodule App.Handler do
  def health, do: "ok"

  def run(path) do
    health()
  end
end

defmodule App.Store do
  def save(key), do: key
end
