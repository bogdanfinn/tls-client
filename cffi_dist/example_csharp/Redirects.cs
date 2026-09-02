namespace TlsClientExamples;

internal static class Redirects
{
    private const string SessionId = "my-redirect-session";

    internal static void Run()
    {
        var payload = new Dictionary<string, object>
        {
            ["tlsClientIdentifier"] = "chrome_150",
            ["followRedirects"] = false,
            ["sessionId"] = SessionId,
            ["proxyUrl"] = "",
            ["headers"] = new Dictionary<string, string> { ["accept"] = "*/*" },
            ["headerOrder"] = new[] { "accept" },
            ["requestUrl"] = "https://httpbin.org/redirect/1",
            ["requestMethod"] = "GET",
            ["requestBody"] = "",
            ["requestCookies"] = Array.Empty<object>()
        };

        var response = TlsClientLibrary.Request(payload);
        Console.WriteLine($"followRedirects=false: {response["status"]} {response["target"]}");
        TlsClientLibrary.FreeMemory((string)response["id"]);

        // followRedirects can be changed within an existing session.
        payload["followRedirects"] = true;

        response = TlsClientLibrary.Request(payload);
        Console.WriteLine($"followRedirects=true: {response["status"]} {response["target"]}");
        TlsClientLibrary.FreeMemory((string)response["id"]);

        TlsClientLibrary.DestroySession(new { sessionId = SessionId });
    }
}
