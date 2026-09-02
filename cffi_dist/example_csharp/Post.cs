namespace TlsClientExamples;

internal static class Post
{
    internal static void Run()
    {
        var response = TlsClientLibrary.Request(new
        {
            tlsClientIdentifier = "chrome_150",
            followRedirects = true,
            proxyUrl = "",
            headers = new Dictionary<string, string>
            {
                ["accept"] = "application/json",
                ["content-type"] = "application/json"
            },
            headerOrder = new[] { "accept", "content-type" },
            requestUrl = "https://httpbin.org/post",
            requestMethod = "POST",
            requestBody = "{\"foo\":\"bar\"}",
            requestCookies = Array.Empty<object>()
        });

        Console.WriteLine($"status: {response["status"]}");
        Console.WriteLine(response["body"]);

        TlsClientLibrary.FreeMemory((string)response["id"]);
    }
}
