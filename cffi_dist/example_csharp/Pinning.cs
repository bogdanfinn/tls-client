namespace TlsClientExamples;

/// <summary>
/// Pins are the base64 encoded SHA-256 hashes of a host's public keys.
/// Generate them with: hpkp-pins -server=bstn.com:443
/// </summary>
internal static class Pinning
{
    internal static void Run()
    {
        var response = TlsClientLibrary.Request(new
        {
            tlsClientIdentifier = "chrome_150",
            followRedirects = true,
            proxyUrl = "",
            certificatePinningHosts = new Dictionary<string, string[]>
            {
                ["bstn.com"] = new[]
                {
                    "NQvy9sFS99nBqk/nZCUF44hFhshrkvxqYtfrZq3i+Ww=",
                    "4a6cPehI7OG6cuDZka5NDZ7FR8a60d3auda+sKfg4Ng=",
                    "x4QzPSC810K5/cMjb05Qm4k3Bw5zBn4lTdO/nEW/Td4="
                }
            },
            headers = new Dictionary<string, string> { ["accept"] = "*/*" },
            headerOrder = new[] { "accept" },
            requestUrl = "https://bstn.com",
            requestMethod = "GET",
            requestBody = "",
            requestCookies = Array.Empty<object>()
        });

        // If a pin does not match, "status" is 0 and "body" holds the pinning
        // error instead of a remote response.
        Console.WriteLine(response);

        TlsClientLibrary.FreeMemory((string)response["id"]);
    }
}
