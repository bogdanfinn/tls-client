namespace TlsClientExamples;

internal static class Cookies
{
    private const string SessionId = "my-cookie-session";

    internal static void Run()
    {
        // The server sets a cookie on this request; the session's cookie jar
        // stores it automatically and replays it on every later request in
        // this session.
        var response = TlsClientLibrary.Request(new
        {
            tlsClientIdentifier = "chrome_150",
            followRedirects = true,
            sessionId = SessionId,
            proxyUrl = "",
            headers = new Dictionary<string, string> { ["accept"] = "*/*" },
            headerOrder = new[] { "accept" },
            requestUrl = "https://httpbin.org/cookies/set?session=abc123",
            requestMethod = "GET",
            requestBody = "",
            requestCookies = Array.Empty<object>()
        });

        TlsClientLibrary.FreeMemory((string)response["id"]);

        var cookies = TlsClientLibrary.GetCookiesFromSession(new
        {
            sessionId = SessionId,
            url = "https://httpbin.org"
        });

        Console.WriteLine(cookies);

        TlsClientLibrary.DestroySession(new { sessionId = SessionId });
    }
}
