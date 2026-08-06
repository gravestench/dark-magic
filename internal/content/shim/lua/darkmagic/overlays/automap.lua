local render=require("dm.render/v1"); local input=require("dm.input/v1"); local scenes=require("dm.scene/v1")
return {blocks_update_below=false,create=function(self) self.root=render.create("hud");self.root:set_position(400,300);self.root:fill_rect(700,500,40,36,28,120) end,update=function(self) if input.pressed("automap") or input.pressed("cancel") then scenes.pop() end end}
