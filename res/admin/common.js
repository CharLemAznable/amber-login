layui.use(['jquery', 'layer', 'form'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;

    $('.admin-header .layui-nav-item a').each(function () {
        let href = $(this).attr('href');
        let path = location.pathname;
        if (path.startsWith(href)) $(this).parent().addClass('layui-this');
    });

    $('#change-password').on('click', function () {
        layer.open({
            type: 1,
            title: '修改密码',
            content:
                '<div class="layui-form layui-form-pane change-password-content">' +
                '    <div class="layui-form-item">' +
                '        <label class="layui-form-label">原密码</label>' +
                '        <div class="layui-input-block">' +
                '            <input name="old-password" type="password" lay-verify="required" lay-verType="tips"' +
                '                   placeholder="请输入原密码" autocomplete="off" class="layui-input">' +
                '        </div>' +
                '    </div>' +
                '    <div class="layui-form-item">' +
                '        <label class="layui-form-label">新密码</label>' +
                '        <div class="layui-input-block">' +
                '            <input name="new-password" type="password" lay-verify="required" lay-verType="tips"' +
                '                   placeholder="请输入新密码" autocomplete="off" class="layui-input">' +
                '        </div>' +
                '    </div>' +
                '    <div class="layui-form-item">' +
                '        <label class="layui-form-label">确认密码</label>' +
                '        <div class="layui-input-block">' +
                '            <input name="renew-password" type="password" lay-verify="required" lay-verType="tips"' +
                '                   placeholder="请再次输入新密码" autocomplete="off" class="layui-input">' +
                '        </div>' +
                '    </div>' +
                '    <div class="layui-form-item">' +
                '        <button class="layui-btn layui-btn-normal confirm-btn" lay-submit lay-filter="change-password">确认</button>' +
                '    </div>' +
                '</div>',
            area: '300px',
            btn: [],
            btnAlign: 'c',
            closeBtn: 1,
            shade: 0.6,
            shadeClose: false,
            id: 'change-password-form',
            resize: false,
            move: false,
        });
    });

    form.on('submit(change-password)', function (data) {
        if (data.field['new-password'] !== data.field['renew-password']) {
            layer.alert('两次输入的新密码不相同');
            return;
        }

        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/admin/change-password',
            contentType: 'application/json',
            data: JSON.stringify(data.field),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg);
                    return;
                }
                layer.alert('操作成功', function () {
                    location.replace(location.href);
                });
            },
            error: function () {
                layer.alert('操作失败');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });

    $('#log-out').on('click', function () {
        layer.confirm('退出登录?', function () {
            let load = layer.load(1, {shade: 0.6});
            $.ajax({
                type: 'POST',
                url: '${contextPath}/admin/do-logout',
                dataType: 'json',
                success: function () {
                    location.replace(location.href);
                },
                error: function () {
                    layer.alert('退出异常');
                },
                complete: function () {
                    layer.close(load);
                },
            });
        });
    });
});
