using StardewModdingAPI;

namespace EchoFarm.Mod;

internal sealed class ModConfig
{
    public string CoreUrl { get; set; } = "http://127.0.0.1:18471";
    public SButton RecordKey { get; set; } = SButton.F7;
    public SButton SummonKey { get; set; } = SButton.F8;
}
