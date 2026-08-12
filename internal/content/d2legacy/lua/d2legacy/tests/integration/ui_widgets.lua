for _, name in ipairs({
    "d2legacy.ui.slider",
    "d2legacy.ui.scrollbar",
    "d2legacy.ui.list",
    "d2legacy.ui.tabs",
    "d2legacy.ui.panel",
    "d2legacy.ui.progress_bar",
}) do
    assert(type(require(name)) == "table", name .. " did not return a module")
end
