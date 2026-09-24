using EchoFarm.Bridge.Contracts;
using EchoFarm.Bridge.Runtime;
using Microsoft.Xna.Framework;
using Microsoft.Xna.Framework.Graphics;
using StardewValley;
using StardewValley.Menus;

namespace EchoFarm.Mod;

internal sealed class EchoMemoryOverlay
{
    private IReadOnlyList<string> statusLines = Array.Empty<string>();
    private IReadOnlyList<string> memoryLines = Array.Empty<string>();
    private IReadOnlyList<string> lines = Array.Empty<string>();

    public bool Visible { get; private set; }

    public void Toggle() => Visible = !Visible;

    public void Hide()
    {
        Visible = false;
    }

    public void UpdateStatus(IReadOnlyList<string> status)
    {
        statusLines = status ?? Array.Empty<string>();
        Rebuild();
    }

    public void Update(EchoMemoryView view)
    {
        memoryLines = EchoMemoryPresenter.BuildLines(view);
        Rebuild();
    }

    public void ShowError(string message)
    {
        memoryLines = new[] { "Echo memory offline", message };
        Rebuild();
    }

    private void Rebuild() => lines = Array.AsReadOnly(statusLines.Concat(memoryLines).Take(14).ToArray());

    public void Draw(SpriteBatch spriteBatch)
    {
        if (!Visible || lines.Count == 0 || !Context.IsWorldReady)
            return;

        const int x = 32;
        const int y = 32;
        const int padding = 20;
        const int width = 720;
        int lineHeight = (int)Game1.smallFont.MeasureString("Echo").Y + 6;
        int height = padding * 2 + lineHeight * lines.Count;
        IClickableMenu.drawTextureBox(
            spriteBatch,
            Game1.menuTexture,
            new Rectangle(0, 256, 60, 60),
            x,
            y,
            width,
            height,
            Color.White * 0.92f,
            1f,
            drawShadow: true
        );
        for (int index = 0; index < lines.Count; index++)
        {
            spriteBatch.DrawString(
                Game1.smallFont,
                lines[index],
                new Vector2(x + padding, y + padding + index * lineHeight),
                Game1.textColor
            );
        }
    }
}
