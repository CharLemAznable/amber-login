layui.use(['jquery', 'layer', 'form'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;

    form.verify({
        pass: [
            /^(?![0-9]+$)(?![a-zA-Z]+$)(?![!-/:-@\\[-`]+$)[0-9A-Za-z!-/:-@\\[-`]{10,20}$/,
            '密码必须为10-20位, 必须包含字母数字和特殊字符'
        ]
    });

    layer.open({
        type: 1,
        title: '修改密码',
        content: $('.change-password-content'),
        area: '300px',
        btn: [],
        btnAlign: 'c',
        closeBtn: false,
        shade: 0.6,
        shadeClose: false,
        id: 'change-password-content',
        resize: false,
        move: false,
    });

    $(document).on('keydown', function (event) {
        // noinspection JSDeprecatedSymbols
        if (event.keyCode === 13) {
            if ($(document).find('.layui-layer').length === 1) {
                $('.change-password-btn').click();
            } else {
                layer.close(layer.index);
            }
        }
    });

    form.on('submit(change-password)', function (data) {
        if (data.field['new-password'] !== data.field['renew-password']) {
            layer.alert('两次输入的新密码不相同');
            return;
        }

        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/do-change-password',
            contentType: 'application/json',
            data: JSON.stringify(data.field),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    if ('验证码不存在或已过期' === res.msg) {
                        layer.alert(res.msg, function () {
                            location.replace(location.href);
                        });
                        return;
                    }
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
});
