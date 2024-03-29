layui.use(['jquery', 'layer', 'form'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;

    layer.open({
        type: 1,
        title: '统一登录',
        content: $('.login-content'),
        area: '300px',
        btn: [],
        btnAlign: 'c',
        closeBtn: false,
        shade: 0.6,
        shadeClose: false,
        id: 'login-content',
        resize: false,
        move: false,
    });

    $(document).on('keydown', function (event) {
        // noinspection JSDeprecatedSymbols
        if (event.keyCode === 13) {
            if ($(document).find('.layui-layer').length === 1) {
                $('.login-btn').click();
            } else {
                layer.close(layer.index);
            }
        }
    });

    form.on('submit(login)', function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/do-login',
            contentType: 'application/json',
            data: JSON.stringify(data.field),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    if ('1' === res.refresh) {
                        layer.alert(res.msg, function () {
                            location.replace(location.href);
                        });
                        return;
                    }
                    layer.alert(res.msg);
                    return;
                }
                location.href = res.redirect;
            },
            error: function () {
                layer.alert('登录异常');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });

    $('.register-btn').on('click', function () {
        window.open('${contextPath}/register', '_blank');
    });

    $('.change-password-btn').on('click', function () {
        window.open('${contextPath}/change-password', '_blank');
    });

    $('.captcha-block').on('click', function () {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            url: '${contextPath}/refresh-captcha',
            contentType: 'application/json',
            success: function (res) {
                $('#captcha-id').val(res['captcha-id']);
                $('#captcha').attr("src", res['captcha']);
            },
            error: function () {
                layer.alert('服务异常');
            },
            complete: function () {
                layer.close(load);
            },
        });
    });
});
