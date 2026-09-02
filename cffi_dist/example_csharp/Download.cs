namespace TlsClientExamples;

internal static class Download
{
    internal static void Run()
    {
        var response = TlsClientLibrary.Request(new
        {
            followRedirects = true,
            proxyUrl = "",
            // isByteResponse base64-encodes the response body, which is
            // required for binary payloads like images.
            isByteResponse = true,
            headers = new Dictionary<string, string> { ["accept"] = "*/*" },
            headerOrder = new[] { "accept" },
            requestUrl = "https://avatars.githubusercontent.com/u/17678241?v=4",
            requestMethod = "GET",
            requestBody = "",
            requestCookies = Array.Empty<object>()
        });

        // The body is a data URI (e.g. "data:image/png;base64,...."); drop
        // everything up to and including the first comma before decoding.
        var body = (string)response["body"];
        var base64Data = body.Substring(body.IndexOf(',') + 1);

        var dest = Path.Combine(Path.GetTempPath(), "tls-client-example.jpg");
        File.WriteAllBytes(dest, Convert.FromBase64String(base64Data));

        Console.WriteLine($"status: {response["status"]}, wrote file to: {dest}");

        TlsClientLibrary.FreeMemory((string)response["id"]);
    }
}
