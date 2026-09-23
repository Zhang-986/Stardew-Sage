using System.Collections;
using System.Reflection;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace EchoFarm.Bridge.Contracts;

public abstract class StrictContract
{
    [JsonExtensionData]
    public Dictionary<string, JsonElement>? UnmappedProperties { get; init; }
}

public static class EchoJson
{
    public static JsonSerializerOptions Options { get; } = CreateOptions();

    public static T Deserialize<T>(string json)
    {
        T? value = JsonSerializer.Deserialize<T>(json, Options);
        if (value is null)
            throw new JsonException("JSON value cannot be null.");

        RejectUnknownMembers(value, "$", new HashSet<object>(ReferenceEqualityComparer.Instance));
        return value;
    }

    public static string Serialize<T>(T value) => JsonSerializer.Serialize(value, Options);

    private static JsonSerializerOptions CreateOptions()
    {
        var options = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
            PropertyNameCaseInsensitive = false,
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
        };
        options.Converters.Add(new JsonStringEnumConverter(new SnakeCaseNamingPolicy(), allowIntegerValues: false));
        return options;
    }

    private static void RejectUnknownMembers(object? value, string path, HashSet<object> visited)
    {
        if (value is null || value is string || value is JsonElement || value.GetType().IsValueType)
            return;
        if (!visited.Add(value))
            return;

        if (value is StrictContract contract && contract.UnmappedProperties is { Count: > 0 })
        {
            string name = contract.UnmappedProperties.Keys.First();
            throw new JsonException($"Unknown JSON property '{name}' at {path}.");
        }

        if (value is IEnumerable items)
        {
            int index = 0;
            foreach (object? item in items)
                RejectUnknownMembers(item, $"{path}[{index++}]", visited);
            return;
        }

        foreach (PropertyInfo property in value.GetType().GetProperties(BindingFlags.Instance | BindingFlags.Public))
        {
            if (property.Name == nameof(StrictContract.UnmappedProperties) || property.GetIndexParameters().Length > 0)
                continue;
            RejectUnknownMembers(property.GetValue(value), $"{path}.{property.Name}", visited);
        }
    }
}

internal sealed class SnakeCaseNamingPolicy : JsonNamingPolicy
{
    public override string ConvertName(string name)
    {
        if (string.IsNullOrEmpty(name))
            return name;

        var result = new StringBuilder(name.Length + 4);
        for (int i = 0; i < name.Length; i++)
        {
            char current = name[i];
            if (char.IsUpper(current) && i > 0)
                result.Append('_');
            result.Append(char.ToLowerInvariant(current));
        }
        return result.ToString();
    }
}
