using StardewModdingAPI;

namespace EchoFarm.Mod;

internal sealed class ModConfig
{
    public string CoreUrl { get; set; } = "http://127.0.0.1:18471";
    public bool AutoStartCore { get; set; } = true;
    public string? CoreExecutablePath { get; set; }
    public int CoreStartupTimeoutSeconds { get; set; } = 10;
    public string ModelMode { get; set; } = "fixture";
    public string? ModelBaseUrl { get; set; }
    public string? ModelName { get; set; }
    public string? DatabasePath { get; set; }
    public bool EnableExperimentalHarvest { get; set; } = false;
    public int MaxModelCallsPerSession { get; set; } = 32;
    public int MaxReportedTokensPerSession { get; set; } = 100000;
    public SButton RecordKey { get; set; } = SButton.F7;
    public SButton SummonKey { get; set; } = SButton.F8;
    public SButton MemoryKey { get; set; } = SButton.F9;
    public SButton CorrectionKey { get; set; } = SButton.F10;
}
