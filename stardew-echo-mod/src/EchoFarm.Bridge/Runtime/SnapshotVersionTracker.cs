namespace EchoFarm.Bridge.Runtime;

public sealed class SnapshotVersionTracker
{
    private readonly object sync = new();
    private string? currentFingerprint;
    private long currentVersion;

    public long Next(string fingerprint)
    {
        ArgumentNullException.ThrowIfNull(fingerprint);
        lock (sync)
        {
            if (!string.Equals(currentFingerprint, fingerprint, StringComparison.Ordinal))
            {
                currentFingerprint = fingerprint;
                currentVersion++;
            }
            return currentVersion;
        }
    }
}
