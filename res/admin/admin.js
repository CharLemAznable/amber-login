layui.use(['jquery', 'layer', 'form'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;

    form.on('submit(admin-submit)', function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/admin/submit-admin',
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
                layer.alert('操作异常');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });
});
