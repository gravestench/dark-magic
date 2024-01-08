local Terminal = {
    log = function(...) print(...) end, -- set by the mod loader, has mod metadata fields
    root = {}, -- the root renderable, acts as a container of ui elements
    box = {}, -- a ui element for the window, child of root
    currentTween = nil, -- used for canceling tween if toggling rapidly
    tweenIn = nil,
    tweenOut = nil,
    enabled = false,
    gateInput = true, -- whether the Terminal should ignore user input
}

function Terminal:Init()
    self.root = api.renderer.NewRenderable()
    self.root.Opacity(0.0) -- during init we hide everything

    self.log("root node", "uuid", self.root.UUID())

    self:InitWindow()
    --self.InitInput()
    --
    --api.events.On("Toggle Terminal", self.Toggle)
    --
    --self.root.SetOpacity(1) -- when we are done, we show everything
end

function Terminal:InitWindow()
    self.log("creating Terminal window")
    w, h = api.renderer.window.Size()

    self.root.Position(0, 0)
    self.root.Enable(false)

    self.box = api.ui.FillRect(0, 0, w, h, 1, "0x333333", "0xefefef")
    self.box.Parent(self.root)

    self.box.ZIndex(-1)

    self.log("setting up tweens")
    self.tweenIn = api.tweens.New()
    self.tweenIn.Ease("Elastic.easeInOut", 0.5, 0.5, 0.5)
    self.tweenIn.OnUpdate(self.onTweenIn)
    self.tweenIn.Time(5*1e9)
    self.currentTween = self.tweenOut

    self.tweenOut = api.tweens.New()
    self.tweenOut.OnUpdate(self.onTweenOut)

    self:Toggle()
end

function Terminal:onTweenIn(progress)
    _, currentY = self.root.Position()
    _, h = api.renderer.window.Size()
    y = -h + (progress * (h - currentY))
    x, _ = self.box.Position()
    self.root.Opacity(progress)
    self.box.Position(x, y)
    self.log("tweening in", "progress", progress, "x", x, "y", y)
end

function Terminal:onTweenOut (progress)
    _, h = api.renderer.window.Size()
    y = -h * progress
    x, _ = self.box.Position()
    self.log("tweening out", "progress", progress, "x", x, "y", y)
    self.box.SetTranslation(x, y)
end

function Terminal:Toggle()
    self.log("toggling")

    self.enabled = not self.enabled
    self.gateInput = not self.enabled

    if self.currentTween ~= nil then
        self.currentTween.Stop()
    end

    if self.enabled then
        self.log("enabling")
        self.gateInput = false
        self.currentTween = self.tweenIn
    else
        self.log("disabling")
        self.gateInput = true
        self.currentTween = self.tweenOut
    end

    self.currentTween.Ease("Elastic.easeInOut", 0.5, 0.5, 0.5)
    self.currentTween.Time(1e15*5)
    self.currentTween.Play()
end

function Terminal:InitInput()
    api.events.On("keypress", self.OnKeyPress)
end

function Terminal:OnKeyPress(keyCode, state)
    if self.gateInput then return end -- ignore if we're gating input

    -- TODO: handle keypress here
end

return Terminal