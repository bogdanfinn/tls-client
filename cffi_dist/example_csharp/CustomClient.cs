namespace TlsClientExamples;

/// <summary>Builds a fingerprint from a raw JA3 string instead of picking a bundled tlsClientIdentifier.</summary>
internal static class CustomClient
{
    internal static void Run()
    {
        var response = TlsClientLibrary.Request(new
        {
            followRedirects = true,
            proxyUrl = "",
            customTlsClient = new
            {
                ja3String = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-10-11-13-16-23-43-51-65281-45-21,29-23-24,0",
                h2Settings = new Dictionary<string, int>
                {
                    ["HEADER_TABLE_SIZE"] = 65536,
                    ["MAX_CONCURRENT_STREAMS"] = 1000,
                    ["INITIAL_WINDOW_SIZE"] = 6291456,
                    ["MAX_HEADER_LIST_SIZE"] = 262144
                },
                h2SettingsOrder = new[] { "HEADER_TABLE_SIZE", "MAX_CONCURRENT_STREAMS", "INITIAL_WINDOW_SIZE", "MAX_HEADER_LIST_SIZE" },
                supportedSignatureAlgorithms = new[] { "ECDSAWithP256AndSHA256", "PSSWithSHA256", "PKCS1WithSHA256" },
                supportedVersions = new[] { "GREASE", "1.3", "1.2" },
                keyShareCurves = new[] { "GREASE", "X25519" },
                certCompressionAlgos = new[] { "brotli" },
                alpnProtocols = new[] { "h2", "http/1.1" },
                alpsProtocols = new[] { "h2" },
                pseudoHeaderOrder = new[] { ":method", ":authority", ":scheme", ":path" },
                connectionFlow = 15663105
            },
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
