function tabs(n)
    if n < 1 then return "" end

    str = ""

    for i = 0, n-1 do
        str = str.."\t"
    end

    return str
end

function tree(obj, prefix)
    if not prefix then
        prefix = ""
    end

    for k,v in pairs(obj) do
        local valueType = type(v)

        if valueType == "table" then
            --print(prefix .. "." .. k)
            tree(v, prefix .. "." .. k)
        elseif valueType == "function" then
            print(prefix .. "." .. k .."()")
        end
    end
end

tree(example, "Example")