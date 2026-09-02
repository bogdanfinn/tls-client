namespace TlsClientExamples;

internal static class Proxy
{
    private const string SessionId = "my-proxy-session";

    internal static void Run()
    {
        var payload = new Dictionary<string, object>
        {
            ["tlsClientIdentifier"] = "chrome_150",
            ["followRedirects"] = false,
            ["sessionId"] = SessionId,
            // http://, socks5:// and socks5h:// proxy URLs are all supported.
            ["proxyUrl"] = "http://user:pass@proxy-one.example.com:8000",
            ["headers"] = new Dictionary<string, string> { ["accept"] = "*/*" },
            ["headerOrder"] = new[] { "accept" },
            ["requestUrl"] = "https://tls.peet.ws/api/all",
            ["requestMethod"] = "GET",
            ["requestBody"] = "",
            ["requestCookies"] = Array.Empty<object>()
        };

        var response = TlsClientLibrary.Request(payload);
        Console.WriteLine($"via proxy one: {response["status"]}");
        TlsClientLibrary.FreeMemory((string)response["id"]);

        // Swap the proxy for the same session - no new client/session needs to be built.
        payload["proxyUrl"] = "http://user:pass@proxy-two.example.com:8000";

        response = TlsClientLibrary.Request(payload);
        Console.WriteLine($"via proxy two: {response["status"]}");
        TlsClientLibrary.FreeMemory((string)response["id"]);

        TlsClientLibrary.DestroySession(new { sessionId = SessionId });
    }
}
