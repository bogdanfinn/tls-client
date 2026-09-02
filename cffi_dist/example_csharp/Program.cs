using TlsClientExamples;

var topics = new Dictionary<string, Action>(StringComparer.OrdinalIgnoreCase)
{
    ["basic"] = Basic.Run,
    ["post"] = Post.Run,
    ["cookies"] = Cookies.Run,
    ["proxy"] = Proxy.Run,
    ["redirects"] = Redirects.Run,
    ["custom_client"] = CustomClient.Run,
    ["pinning"] = Pinning.Run,
    ["download"] = Download.Run
};

var topic = args.Length > 0 ? args[0] : "basic";

if (!topics.TryGetValue(topic, out var run))
{
    Console.WriteLine($"unknown topic '{topic}'. available: {string.Join(", ", topics.Keys)}");
    return;
}

// dotnet run -- <topic>, e.g. `dotnet run -- cookies`
run();
