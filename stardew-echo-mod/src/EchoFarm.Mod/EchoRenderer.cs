using Microsoft.Xna.Framework;
using Microsoft.Xna.Framework.Graphics;
using StardewModdingAPI;
using StardewValley;

namespace EchoFarm.Mod;

internal sealed class EchoRenderer
{
    private readonly EchoAvatarState echo;

    public EchoRenderer(EchoAvatarState echo)
    {
        this.echo = echo;
    }

    public void Draw(SpriteBatch spriteBatch)
    {
        if (!echo.Visible || !Context.IsWorldReady)
            return;

        Farmer source = Game1.player;
        Vector2 worldPosition = echo.Tile * Game1.tileSize;
        Vector2 screenPosition = Game1.GlobalToLocal(Game1.viewport, worldPosition);
        float layerDepth = Math.Clamp((worldPosition.Y + Game1.tileSize) / 10000f, 0f, 1f);
        source.FarmerRenderer.draw(
            spriteBatch,
            source.FarmerSprite,
            source.FarmerSprite.SourceRect,
            screenPosition,
            Vector2.Zero,
            layerDepth,
            Color.Cyan * 0.55f,
            0f,
            source
        );
    }
}
