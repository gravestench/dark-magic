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

    self.log("root node", "uuid", self.root.UUID())

    self:initGraphics()
    --self.InitInput()
    --
    --api.events.On("Toggle Terminal", self.Toggle)
    --
    --self.root.SetOpacity(1) -- when we are done, we show everything
end

function Terminal:initGraphics()
    self.log("setting up terminal graphics")
    w, windowHeight = api.renderer.window.Size()

    self.root.Position(0, 0)
    self.root.Enable(false)

    self.box = api.ui.FillRect(0, 0, w, windowHeight, 1, "0x343434", "0x787878")
    self.box.Position(0, -windowHeight)
    self.box.Origin(0, 0)
    self.box.Parent(self.root)
    self.box.ZIndex(-1)

    self:initTweens()
end

function Terminal:initTweens()
    self.log("setting up tweens")

    self.tweenIn = api.tweens.New()
    self.tweenOut = api.tweens.New()

    second = 1e9
    tweenTime = 10 * second
    tweenEase = "Linear"

    self.log("setting tween times", "seconds", tweenTime/second)
    self.tweenIn.Time(tweenTime)
    self.tweenOut.Time(tweenTime)

    self.log("setting tween ease", "ease", tweenEase)
    self.tweenIn.Ease("Linear")
    self.tweenOut.Ease("Linear")

    self.currentTween = self.tweenOut

    self.log("setting tween callbacks")
    self.tweenIn.OnUpdate(function (progress) self:onTweenIn(progress) end)
    self.tweenOut.OnUpdate(function (progress) self:onTweenOut(progress) end)

    self.log("stopping tweens")
    self.tweenIn.Stop()
    self.tweenOut.Stop()

    self.log("adding tweens to tween manager")
    api.tweens.Add(self.tweenIn)
    api.tweens.Add(self.tweenOut)

    self.log("debug test toggling to animate terminal in")
    self:Toggle()
end

function Terminal:onTweenIn(progress)
    windowWidth, windowHeight = api.renderer.window.Size()

    startX, startY = self.root.Position()
    targetX, targetY = startX, 0

    offset = progress * (windowHeight - startY)
    y = windowHeight - offset

    x, _ = self.box.Position()

    self.root.Opacity(progress)
    self.box.Position(x, y)
end

function Terminal:onTweenOut(progress)
    _, windowHeight = api.renderer.window.Size()
    offset = windowHeight * progress
    y = windowHeight - offset
    x, _ = self.box.Position()
    self.root.Opacity(1-progress)
    self.box.Position(x, y)
end

function Terminal:Toggle()
    self.enabled = not self.enabled
    self.gateInput = not self.enabled

    if self.enabled then
        self.log("toggling", "enabled", self.enabled, "gate input", self.gateInput, "tween", "tween in")
        self.gateInput = false
        self.tweenOut.Stop()
        self.tweenIn.Start()
    else
        self.log("toggling", "enabled", self.enabled, "gate input", self.gateInput, "tween", "tween out")
        self.gateInput = true
        self.tweenIn.Stop()
        self.tweenOut.Start()
    end

end

function Terminal:InitInput()
    api.events.On("keypress", self.OnKeyPress)
end

function Terminal:OnKeyPress(keyCode, state)
    if self.gateInput then return end -- ignore if we're gating input

    -- TODO: handle keypress here
end

return Terminal