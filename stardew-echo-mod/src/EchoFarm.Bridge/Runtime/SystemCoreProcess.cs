using System.Diagnostics;

namespace EchoFarm.Bridge.Runtime;

public sealed class HttpCoreHealthProbe : ICoreHealthProbe
{
    private readonly HttpClient client;
    private readonly TimeSpan timeout;

    public HttpCoreHealthProbe(HttpClient client, TimeSpan timeout)
    {
        this.client = client ?? throw new ArgumentNullException(nameof(client));
        if (timeout <= TimeSpan.Zero)
            throw new ArgumentOutOfRangeException(nameof(timeout));
        this.timeout = timeout;
    }

    public async Task<bool> IsHealthyAsync(Uri baseUri, CancellationToken cancellationToken)
    {
        using var timeoutSource = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeoutSource.CancelAfter(timeout);
        try
        {
            using HttpResponseMessage response = await client.GetAsync(
                new Uri(baseUri, "/healthz"),
                timeoutSource.Token
            ).ConfigureAwait(false);
            return response.IsSuccessStatusCode;
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            return false;
        }
        catch (HttpRequestException)
        {
            return false;
        }
    }
}

public sealed class SystemCoreProcessLauncher : ICoreProcessLauncher
{
    private readonly Action<string, bool>? log;

    public SystemCoreProcessLauncher(Action<string, bool>? log = null)
    {
        this.log = log;
    }

    public ICoreProcess Start(CoreProcessStartInfo startInfo)
    {
        if (!File.Exists(startInfo.ExecutablePath))
            throw new CoreUnavailableException($"Bundled EchoFarm core was not found at '{startInfo.ExecutablePath}'.");
        if (!Directory.Exists(startInfo.WorkingDirectory))
            Directory.CreateDirectory(startInfo.WorkingDirectory);
        if (!OperatingSystem.IsWindows())
            EnsureUnixExecutable(startInfo.ExecutablePath);

        var process = new Process
        {
            StartInfo = new ProcessStartInfo
            {
                FileName = startInfo.ExecutablePath,
                WorkingDirectory = startInfo.WorkingDirectory,
                UseShellExecute = false,
                CreateNoWindow = true,
                RedirectStandardOutput = true,
                RedirectStandardError = true
            }
        };
        foreach ((string key, string value) in startInfo.Environment)
            process.StartInfo.Environment[key] = value;
        process.OutputDataReceived += (_, args) => WriteLog(args.Data, isError: false);
        process.ErrorDataReceived += (_, args) => WriteLog(args.Data, isError: true);

        try
        {
            if (!process.Start())
                throw new CoreUnavailableException("The bundled EchoFarm core process could not be started.");
            process.BeginOutputReadLine();
            process.BeginErrorReadLine();
            return new SystemCoreProcess(process);
        }
        catch (CoreUnavailableException)
        {
            process.Dispose();
            throw;
        }
        catch (Exception error)
        {
            process.Dispose();
            throw new CoreUnavailableException("The bundled EchoFarm core process could not be started.", error);
        }
    }

    private void WriteLog(string? message, bool isError)
    {
        if (!string.IsNullOrWhiteSpace(message))
            log?.Invoke(message, isError);
    }

    private static void EnsureUnixExecutable(string executablePath)
    {
        using var chmod = new Process
        {
            StartInfo = new ProcessStartInfo
            {
                FileName = "/bin/chmod",
                UseShellExecute = false,
                CreateNoWindow = true
            }
        };
        chmod.StartInfo.ArgumentList.Add("u+x");
        chmod.StartInfo.ArgumentList.Add(executablePath);
        if (!chmod.Start())
            throw new CoreUnavailableException("Could not set execute permission on the bundled EchoFarm core.");
        chmod.WaitForExit();
        if (chmod.ExitCode != 0)
            throw new CoreUnavailableException("Could not set execute permission on the bundled EchoFarm core.");
    }
}

public sealed class TaskAsyncDelay : IAsyncDelay
{
    public Task WaitAsync(TimeSpan delay, CancellationToken cancellationToken) =>
        Task.Delay(delay, cancellationToken);
}

internal sealed class SystemCoreProcess : ICoreProcess
{
    private readonly Process process;

    public SystemCoreProcess(Process process)
    {
        this.process = process;
    }

    public bool HasExited => process.HasExited;
    public int? ExitCode => process.HasExited ? process.ExitCode : null;

    public void Terminate()
    {
        if (process.HasExited)
            return;
        process.Kill(entireProcessTree: true);
        process.WaitForExit(5000);
    }

    public void Dispose() => process.Dispose();
}
