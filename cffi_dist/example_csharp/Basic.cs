namespace TlsClientExamples;

/// <summary>The smallest possible request: one GET call, no session reuse.</summary>
internal static class Basic
{
    internal static void Run()
    {
        var response = TlsClientLibrary.Request(new
        {
            tlsClientIdentifier = "chrome_150",
            followRedirects = true,
            proxyUrl = "",
            headers = new Dictionary<string, string> { ["accept"] = "*/*" },
            headerOrder = new[] { "accept" },
            requestUrl = "https://tls.peet.ws/api/all",
            requestMethod = "GET",
            requestBody = "",
            requestCookies = Array.Empty<object>()
        });

        Console.WriteLine($"status: {response["status"]}");
        Console.WriteLine(response["body"]);

        TlsClientLibrary.FreeMemory((string)response["id"]);
    }
}
