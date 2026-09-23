namespace EchoFarm.Bridge.Transport;

public class EchoFarmException : Exception
{
    public EchoFarmException(string message) : base(message) { }

    public EchoFarmException(string message, Exception innerException) : base(message, innerException) { }
}

public sealed class ModelUnavailableException : EchoFarmException
{
    public ModelUnavailableException() : base("EchoFarm model is unavailable.") { }
}

public class EchoFarmProtocolException : EchoFarmException
{
    public EchoFarmProtocolException(string message) : base(message) { }

    public EchoFarmProtocolException(string message, Exception innerException) : base(message, innerException) { }
}

public sealed class StaleActionException : EchoFarmProtocolException
{
    public StaleActionException() : base("EchoFarm returned an action for a stale world snapshot.") { }
}

public sealed class EchoMemoryNotFoundException : EchoFarmException
{
    public EchoMemoryNotFoundException() : base("No Echo memory exists for this save.") { }
}
