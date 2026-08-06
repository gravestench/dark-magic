local render=require("dm.render/v1"); local input=require("dm.input/v1"); local scenes=require("dm.scene/v1")
return {blocks_update_below=true,create=function(self) self.root=render.create("modal");self.root:set_position(590,300);self.root:fill_rect(380,500,22,16,10,245) end,update=function(self) if input.pressed("skills") or input.pressed("cancel") then scenes.pop() end end}
