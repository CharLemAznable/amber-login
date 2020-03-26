layui.use(['jquery', 'layer', 'form'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;

    let load = layer.load(1, {shade: 0.6});
    $.ajax({
        url: '${contextPath}/admin/query-apps',
        dataType: 'json',
        success: function (res) {
            if ('OK' !== res.msg) {
                layer.alert(res.msg);
                return;
            }

            let appsHtml = '<div class="layui-row layui-col-space15 apps-container-row">';
            let len = res['apps'].length;
            for (let i = 0; i < len; i++) {
                let app = res['apps'][i];
                appsHtml +=
                    '<div class="layui-col-lg4 layui-col-md4 layui-col-sm4 layui-col-xs4">' +
                    '    <div class="layui-card" id="' + app.id + '">' +
                    '        <div class="layui-card-header">' +
                    '            <label class="app-name">' + app.name + '[' + app.id + ']</label>' +
                    '            <div class="layui-btn-group app-action">' +
                    '                <button type="button" class="layui-btn layui-btn-xs app-copy">复制</button>' +
                    '                <button type="button" class="layui-btn layui-btn-xs app-edit">编辑</button>' +
                    '                <button type="button" class="layui-btn layui-btn-xs layui-btn-danger app-delete">删除</button>' +
                    '            </div>' +
                    '        </div>' +
                    '        <div class="layui-card-body">' +
                    '            <div><label class="app-attr">CookieDomain:</label>' +
                    '                <label>' + app['cookie-domain'] + '</label></div>' +
                    '            <div><label class="app-attr">CookieName:</label>' +
                    '                <label>' + app['cookie-name'] + '</label></div>' +
                    '            <div><label class="app-attr">AES密钥:</label>' +
                    '                <label>' + app['encrypt-key'] + '</label></div>' +
                    '            <div><label class="app-attr">默认地址:</label>' +
                    '                <label>' + app['default-url'] + '</label></div>' +
                    '            <div><label class="app-attr">跨域回调地址:</label>' +
                    '                <label>' + app['cocs-url'] + '</label></div>' +
                    '        </div>' +
                    '    </div>' +
                    '</div>';
                if (i % 3 === 2) {
                    appsHtml += '</div><div class="layui-row layui-col-space15 apps-container-row">';
                }
            }
            appsHtml +=
                '<div class="layui-col-lg4 layui-col-md4 layui-col-sm4 layui-col-xs4">' +
                '    <div class="layui-card layui-card-add">' +
                '        <i class="layui-icon">&#xe654;</i>' +
                '    </div>' +
                '</div>';
            appsHtml += '</div>';
            $('.apps-container').html(appsHtml);
        },
        error: function () {
            layer.alert('查询失败');
        },
        complete: function () {
            layer.close(load);
        },
    });

    let $apps = $('.apps-container');

    $apps.on('click', '.layui-card-add', function () {
        form.val('app-form', {
            'id': '',
            'name': '',
            'cookie-domain': '',
            'cookie-name': '',
            'encrypt-key': '',
        });
        layer.open({
            type: 1,
            title: '添加应用',
            content: $('.app-content'),
            area: '350px',
            btn: [],
            btnAlign: 'c',
            closeBtn: 1,
            shade: 0.6,
            shadeClose: false,
            id: 'submit-app-form',
            resize: false,
            move: false,
        });
    });

    $apps.on('click', '.app-copy', function () {
        let appId = $(this).parents('.layui-card').attr('id');
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            url: '${contextPath}/admin/query-app?appId=' + appId,
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg);
                    return;
                }

                let app = res['app'];
                app.id = '';
                form.val('app-form', app);
                layer.open({
                    type: 1,
                    title: '复制应用',
                    content: $('.app-content'),
                    area: '350px',
                    btn: [],
                    btnAlign: 'c',
                    closeBtn: 1,
                    shade: 0.6,
                    shadeClose: false,
                    id: 'submit-app-form',
                    resize: false,
                    move: false,
                });
            },
            error: function () {
                layer.alert('操作失败');
            },
            complete: function () {
                layer.close(load);
            },
        });
    });

    $apps.on('click', '.app-edit', function () {
        let appId = $(this).parents('.layui-card').attr('id');
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            url: '${contextPath}/admin/query-app?appId=' + appId,
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg);
                    return;
                }

                form.val('app-form', res['app']);
                layer.open({
                    type: 1,
                    title: '编辑应用',
                    content: $('.app-content'),
                    area: '350px',
                    btn: [],
                    btnAlign: 'c',
                    closeBtn: 1,
                    shade: 0.6,
                    shadeClose: false,
                    id: 'submit-app-form',
                    resize: false,
                    move: false,
                });
            },
            error: function () {
                layer.alert('操作失败');
            },
            complete: function () {
                layer.close(load);
            },
        });
    });

    form.on('submit(app-submit)', function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/admin/submit-app',
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

    $apps.on('click', '.app-delete', function () {
        let appId = $(this).parents('.layui-card').attr('id');
        layer.confirm('应用删除后将无法恢复, 确认删除?', {title: '删除应用'}, function () {
            let load = layer.load(1, {shade: 0.6});
            $.ajax({
                type: 'POST',
                url: '${contextPath}/admin/delete-app',
                contentType: 'application/json',
                data: JSON.stringify({id: appId}),
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
        });
    });
});
