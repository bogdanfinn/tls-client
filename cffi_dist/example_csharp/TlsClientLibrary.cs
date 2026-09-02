using System.Runtime.InteropServices;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using Newtonsoft.Json.Serialization;

namespace TlsClientExamples;

/// <summary>
/// Thin wrapper around the tls-client shared library's C ABI. Point
/// LibraryPath at the shared library you downloaded from
/// https://github.com/bogdanfinn/tls-client/releases (or built yourself) -
/// the exact file name depends on your OS, architecture and version.
/// </summary>
internal static class TlsClientLibrary
{
    private const string LibraryPath = "../dist/tls-client-xgo-1.16.0-linux-amd64.so";

    private static readonly JsonSerializerSettings SerializerSettings = new()
    {
        ContractResolver = new CamelCasePropertyNamesContractResolver()
    };

    [DllImport(LibraryPath, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr request(string payload);

    [DllImport(LibraryPath, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr getCookiesFromSession(string payload);

    [DllImport(LibraryPath, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr destroySession(string payload);

    [DllImport(LibraryPath, CallingConvention = CallingConvention.Cdecl)]
    private static extern void freeMemory(string responseId);

    internal static JObject Request(object payload) => Call(request, payload);

    internal static JObject GetCookiesFromSession(object payload) => Call(getCookiesFromSession, payload);

    internal static JObject DestroySession(object payload) => Call(destroySession, payload);

    /// <summary>Releases the memory held by a previous response. Call this once you are done reading it.</summary>
    internal static void FreeMemory(string responseId) => freeMemory(responseId);

    private static JObject Call(Func<string, IntPtr> invoke, object payload)
    {
        var payloadJson = JsonConvert.SerializeObject(payload, SerializerSettings);
        var responsePtr = invoke(payloadJson);

        return JObject.Parse(Marshal.PtrToStringAnsi(responsePtr) ?? "{}");
    }
}
