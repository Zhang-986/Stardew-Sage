using EchoFarm.Bridge.Runtime;
using Microsoft.Xna.Framework;

namespace EchoFarm.Mod;

internal sealed class EchoAvatarState
{
    public Vector2 Tile { get; set; }
    public int Energy { get; set; } = 270;
    public int Water { get; set; } = 40;
    public int WaterCapacity { get; set; } = 40;
    public bool Visible { get; set; }
    public EchoInventory Inventory { get; } = new(capacity: 12);
}
