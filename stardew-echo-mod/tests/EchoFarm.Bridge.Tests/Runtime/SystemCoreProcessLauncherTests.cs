using EchoFarm.Bridge.Runtime;

namespace EchoFarm.Bridge.Tests.Runtime;

public sealed class SystemCoreProcessLauncherTests
{
    [Fact]
    public void UnixLauncherRestoresMissingExecutePermission()
    {
        if (OperatingSystem.IsWindows())
            return;

        string directory = Path.Combine(Path.GetTempPath(), $"echofarm-launcher-{Guid.NewGuid():N}");
        Directory.CreateDirectory(directory);
        string executable = Path.Combine(directory, "echofarm-core");
        File.WriteAllText(executable, "#!/bin/sh\nexit 0\n");
        File.SetUnixFileMode(executable, UnixFileMode.UserRead | UnixFileMode.UserWrite);
        try
        {
            using ICoreProcess process = new SystemCoreProcessLauncher().Start(new CoreProcessStartInfo(
                executable,
                directory,
                new Dictionary<string, string>()
            ));

            Assert.True(SpinWait.SpinUntil(() => process.HasExited, TimeSpan.FromSeconds(10)));
            Assert.Equal(0, process.ExitCode);
            Assert.True(File.GetUnixFileMode(executable).HasFlag(UnixFileMode.UserExecute));
        }
        finally
        {
            Directory.Delete(directory, recursive: true);
        }
    }
}
