using System;

namespace Api;

public class Handler
{
    public string Health() => "ok";

    public void Run()
    {
        Console.WriteLine(Health());
    }
}
