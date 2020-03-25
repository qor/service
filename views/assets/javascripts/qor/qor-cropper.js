(function(factory) {
    if (typeof define === 'function' && define.amd) {
        // AMD. Register as anonymous module.
        define(['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node / CommonJS
        factory(require('jquery'));
    } else {
        // Browser globals.
        factory(jQuery);
    }
})(function($) {
    'use strict';

    let URL = window.URL || window.webkitURL,
        NAMESPACE = 'qor.cropper',
        // Events
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SHOWN = 'shown.qor.modal',
        EVENT_HIDDEN = 'hidden.qor.modal',
        // Classes
        CLASS_TOGGLE = '.qor-cropper__toggle',
        CLASS_CANVAS = '.qor-cropper__canvas',
        CLASS_WRAPPER = '.qor-cropper__wrapper',
        CLASS_OPTIONS = '.qor-cropper__options',
        CLASS_SAVE = '.qor-cropper__save',
        CLASS_DELETE = '.qor-cropper__toggle--delete',
        CLASS_CROP = '.qor-cropper__toggle--crop',
        CLASS_UNDO = '.qor-fieldset__undo',
        HIDDEN_DATA_INPUT = 'input[name="QorResource.MediaOption"]:hidden';

    function capitalize(str) {
        if (typeof str === 'string') {
            str = str.charAt(0).toUpperCase() + str.substr(1);
        }

        return str;
    }

    function getLowerCaseKeyObject(obj) {
        let newObj = {},
            key;

        if ($.isPlainObject(obj)) {
            for (key in obj) {
                if (obj.hasOwnProperty(key)) {
                    newObj[String(key).toLowerCase()] = obj[key];
                }
            }
        }

        return newObj;
    }

    function getValueByNoCaseKey(obj, key) {
        let originalKey = String(key),
            lowerCaseKey = originalKey.toLowerCase(),
            upperCaseKey = originalKey.toUpperCase(),
            capitalizeKey = capitalize(originalKey);

        if ($.isPlainObject(obj)) {
            return obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey];
        }
    }

    function replaceText(str, data) {
        if (typeof str === 'string') {
            if (typeof data === 'object') {
                $.each(data, function(key, val) {
                    str = str.replace('$[' + String(key).toLowerCase() + ']', val);
                });
            }
        }

        return str;
    }

    function isSVG(url) {
        return /.svg$/.test(url);
    }

    function QorCropper(element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorCropper.DEFAULTS, $.isPlainObject(options) && options);
        this.data = null;
        this.init();
    }

    QorCropper.prototype = {
        constructor: QorCropper,

        init: function() {
            let options = this.options,
                $this = this.$element,
                $parent = $this.closest(options.parent),
                data,
                outputValue,
                fetchUrl,
                _this = this,
                imageData;

            if (!$parent.length) {
                $parent = $this.parent();
            }

            this.$parent = $parent;
            this.$output = $parent.find(options.output);
            this.$formCropInput = $parent.closest('form').find(HIDDEN_DATA_INPUT);
            this.$list = $parent.find(options.list);

            fetchUrl = this.$output.data('fetchSizedata');

            if (fetchUrl) {
                $.getJSON(fetchUrl, function(data) {
                    imageData = JSON.parse(data.MediaOption);
                    _this.$output.val(JSON.stringify(data));
                    _this.$formCropInput.val(JSON.stringify(data));
                    _this.data = imageData || {};
                    if (isSVG(imageData.URL || imageData.Url)) {
                        _this.resetImage();
                    }
                    _this.build();
                    _this.bind();
                });
            } else {
                outputValue = $.trim(this.$output.val());
                if (outputValue) {
                    data = JSON.parse(outputValue);
                    if (isSVG(data.URL || data.Url)) {
                        this.resetImage();
                    }
                }

                this.data = data || {};

                this.build();
                this.bind();
            }
        },

        resetImage: function() {
            this.$parent.addClass('is-svg');
        },

        build: function() {
            let textData = this.$output.data(),
                text = {},
                replaceTexts;

            if (textData) {
                text = {
                    title: textData.cropperTitle,
                    ok: textData.cropperOk,
                    cancel: textData.cropperCancel
                };
                replaceTexts = this.options.text;
            }

            if (text.ok && text.title && text.cancel) {
                replaceTexts = text;
            }

            this.wrap();
            this.$modal = $(replaceText(QorCropper.MODAL, replaceTexts)).appendTo('body');
        },

        unbuild: function() {
            this.$modal.remove();
            this.unwrap();
        },

        wrap: function() {
            let $list = this.$list,
                $img;

            $img = $list.find('img').not('.is-svg');

            if ($img.length) {
                $list.find('li').append(QorCropper.TOGGLE);
                $img.wrap(QorCropper.CANVAS);
                this.center($img);
            } else {
                $list.find(CLASS_CROP).remove();
            }
        },

        unwrap: function() {
            let $list = this.$list;

            $list.find(CLASS_TOGGLE).remove();
            $list.find(CLASS_CANVAS).each(function() {
                let $this = $(this);

                $this.before($this.html()).remove();
            });
        },

        bind: function() {
            this.$element.on(EVENT_CHANGE, $.proxy(this.read, this));
            this.$list.on(EVENT_CLICK, $.proxy(this.click, this));
            this.$modal.on(EVENT_SHOWN, $.proxy(this.start, this)).on(EVENT_HIDDEN, $.proxy(this.stop, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CHANGE, this.read);
            this.$list.off(EVENT_CLICK, this.click);
            this.$modal.off(EVENT_SHOWN, this.start).off(EVENT_HIDDEN, this.stop);
        },

        click: function(e) {
            let target = e.target,
                $target,
                data = this.data,
                $alert;

            if (target === this.$list[0]) {
                return;
            }

            $target = $(target);

            if ($target.closest(CLASS_DELETE).length) {
                data.Delete = true;

                this.$output.val(JSON.stringify(data));
                this.$formCropInput.val(JSON.stringify(data));

                this.$list.hide();

                $alert = $(QorCropper.ALERT);
                $alert.find(CLASS_UNDO).one(
                    EVENT_CLICK,
                    function() {
                        $alert.remove();
                        this.$list.show();
                        delete data.Delete;
                        this.$output.val(JSON.stringify(data));
                        this.$formCropInput.val(JSON.stringify(data));
                    }.bind(this)
                );
                this.$parent.find('.qor-fieldset').append($alert);
            }

            if ($target.closest(CLASS_CROP).length) {
                $target = $target.closest('li').find('img');
                this.$target = $target;
                this.$modal.qorModal('show');
            }
        },

        read: function(e) {
            let files = e.target.files,
                file,
                $list = this.$list,
                $alert = this.$parent.find('.qor-fieldset__alert');

            $list.show();

            if ($alert.length) {
                $alert.remove();
            }

            if (files && files.length) {
                file = files[0];

                if (/^image\//.test(file.type) && URL) {
                    this.fileType = file.type;
                    this.load(URL.createObjectURL(file));
                    this.$parent.find('.qor-medialibrary__image-desc').show();
                } else {
                    $list.empty().html(QorCropper.FILE_LIST.replace('{{filename}}', file.name));
                }
            }
        },

        load: function(url, fromExternal, callback) {
            let options = this.options,
                _this = this,
                $list = this.$list,
                $ul = $(QorCropper.LIST),
                data = this.data || {},
                outputOriginalData = this.$output.val() ? JSON.parse(this.$output.val()) : {},
                outputOriginalDataType = ['Video', 'SelectedType', 'Description'],
                fileType = this.fileType,
                $image,
                imageLength;

            // media box will use load method, has it's own html structure.
            if (!fromExternal) {
                $list.find('ul').remove();
                $list.html($ul);
            }

            $image = $list.find('img');
            this.wrap();

            imageLength = $image.length;
            $image
                .one('load', function() {
                    if (fileType === 'image/svg+xml') {
                        $list.find(CLASS_TOGGLE).remove();
                        return false;
                    }

                    let $this = $(this),
                        naturalWidth = this.naturalWidth,
                        naturalHeight = this.naturalHeight,
                        sizeData = $this.data(),
                        sizeResolution = sizeData.sizeResolution,
                        sizeName = sizeData.sizeName,
                        emulateImageData = {},
                        emulateCropData = {},
                        aspectRatio,
                        width = sizeData.sizeResolutionWidth,
                        height = sizeData.sizeResolutionHeight;

                    if (sizeResolution) {
                        if (!width && !height) {
                            width = getValueByNoCaseKey(sizeResolution, 'width');
                            height = getValueByNoCaseKey(sizeResolution, 'height');
                        }
                        if (height && width) {
                            aspectRatio = width / height;
                        } else {
                            aspectRatio = naturalWidth / naturalHeight;
                        }

                        if (naturalHeight * aspectRatio > naturalWidth) {
                            width = naturalWidth;
                            height = width / aspectRatio;
                        } else {
                            height = naturalHeight;
                            width = height * aspectRatio;
                        }

                        emulateImageData = {
                            naturalWidth: naturalWidth,
                            naturalHeight: naturalHeight
                        };

                        emulateCropData = {
                            x: Math.round((naturalWidth - width) / 2),
                            y: Math.round((naturalHeight - height) / 2),
                            width: Math.round(width),
                            height: Math.round(height)
                        };

                        _this.preview($this, emulateImageData, emulateCropData);

                        if (sizeName) {
                            data.Crop = true;

                            if (!data[options.key]) {
                                data[options.key] = {};
                            }

                            if (sizeName != 'original') {
                                data[options.key][sizeName] = emulateCropData;
                            }
                        }
                    } else {
                        _this.center($this);
                    }

                    // Crop, CropOptions and Delete should be BOOL type, if empty should delete,
                    if (data.Crop === '' || !fromExternal) {
                        delete data.Crop;
                    }

                    if (!fromExternal) {
                        data.CropOptions = null;
                        delete data.Sizes;
                    }

                    delete data.Delete;

                    //   Fix media upload image other property missing bug: ["Video", "SelectedType", "Description"]
                    outputOriginalDataType.forEach((type) => {
                        if (outputOriginalData[type]) {
                            data[type] = outputOriginalData[type];
                        }
                    });

                    _this.$output.val(JSON.stringify(data));
                    _this.$formCropInput.val(JSON.stringify(data));

                    // callback after load complete
                    if (sizeName && data[options.key] && Object.keys(data[options.key]).length >= imageLength) {
                        if (callback && $.isFunction(callback)) {
                            callback();
                        }
                    }
                })
                .attr('src', url)
                .data('originalUrl', url);

            $list.show();
        },

        start: function() {
            let options = this.options,
                $modal = this.$modal,
                $target = this.$target,
                sizeData = $target.data(),
                sizeName = sizeData.sizeName || 'original',
                sizeResolution = sizeData.sizeResolution,
                originalUrl = /\.original\./.test(sizeData.originalUrl)
                    ? sizeData.originalUrl
                    : /\.original\./.test($target.attr('data-original-url'))
                    ? $target.attr('data-original-url')
                    : $target.attr('data-original-url').replace(/file\./, 'file.original.'),
                $clone = $(`<img src=${originalUrl}>`),
                data = this.data || {},
                _this = this,
                sizeAspectRatio = NaN,
                sizeWidth = sizeData.sizeResolutionWidth,
                sizeHeight = sizeData.sizeResolutionHeight,
                list;

            if (sizeResolution) {
                if (!sizeWidth && !sizeHeight) {
                    sizeWidth = getValueByNoCaseKey(sizeResolution, 'width');
                    sizeHeight = getValueByNoCaseKey(sizeResolution, 'height');
                }
                sizeAspectRatio = sizeWidth / sizeHeight;
            }

            if (!data[options.key]) {
                data[options.key] = {};
            }

            $modal
                .trigger('enable.qor.material')
                .find(CLASS_WRAPPER)
                .html($clone);

            list = this.getList(sizeAspectRatio);

            if (list) {
                $modal
                    .find(CLASS_OPTIONS)
                    .show()
                    .append(list);
            }

            $clone.cropper({
                aspectRatio: sizeAspectRatio,
                data: getLowerCaseKeyObject(data[options.key][sizeName]),
                background: false,
                movable: false,
                zoomable: false,
                scalable: false,
                rotatable: false,
                autoCropArea: 1,

                ready: function() {
                    $modal
                        .find('.qor-cropper__options-toggle')
                        .on(EVENT_CLICK, function() {
                            $modal.find('.qor-cropper__options-input').prop('checked', $(this).prop('checked'));
                        })
                        .prop('checked', true);

                    $modal.find(CLASS_SAVE).one(EVENT_CLICK, function() {
                        let cropData = $clone.cropper('getData', true),
                            croppedCanvas = $clone.cropper('getCroppedCanvas'),
                            syncData = [],
                            url;

                        data.Crop = true;
                        data[options.key][sizeName] = cropData;
                        _this.imageData = $clone.cropper('getImageData');
                        _this.cropData = cropData;

                        if (croppedCanvas) {
                            try {
                                url = croppedCanvas.toDataURL();
                            } catch (error) {
                                console.log(error);
                                console.log('Please check image Cross-origin setting');
                            }
                        }

                        $modal.find(CLASS_OPTIONS + ' input').each(function() {
                            let $this = $(this);

                            if ($this.prop('checked')) {
                                syncData.push($this.attr('name'));
                            }
                        });

                        _this.output(url, syncData);
                        $modal.qorModal('hide');
                    });
                }
            });
        },

        stop: function() {
            this.$modal
                .trigger('disable.qor.material')
                .find(CLASS_WRAPPER + ' > img')
                .cropper('destroy')
                .remove()
                .end()
                .find(CLASS_OPTIONS)
                .hide()
                .find('ul')
                .remove();
        },

        getList: function(aspectRatio) {
            let list = [];

            this.$list
                .find('img')
                .not(this.$target)
                .each(function() {
                    let data = $(this).data(),
                        resolution = data.sizeResolution,
                        name = data.sizeName,
                        width = data.sizeResolutionWidth,
                        height = data.sizeResolutionHeight;

                    if (resolution) {
                        if (!width && !height) {
                            width = getValueByNoCaseKey(resolution, 'width');
                            height = getValueByNoCaseKey(resolution, 'height');
                        }

                        if (width / height === aspectRatio) {
                            list.push(
                                '<label>' +
                                    '<input class="qor-cropper__options-input" type="checkbox" name="' +
                                    name +
                                    '" checked> ' +
                                    '<span>' +
                                    name +
                                    '<small>(' +
                                    width +
                                    '&times;' +
                                    height +
                                    ' px)</small>' +
                                    '</span>' +
                                    '</label>'
                            );
                        }
                    }
                });

            return list.length ? '<ul><li>' + list.join('</li><li>') + '</li></ul>' : '';
        },

        output: function(url, data) {
            let $target = this.$target;

            if (url) {
                this.center($target.attr('src', url), true);
            } else {
                this.preview($target);
            }

            if ($.isArray(data) && data.length) {
                this.autoCrop(url, data);
            }

            this.$output.val(JSON.stringify(this.data)).trigger(EVENT_CHANGE);
            this.$formCropInput.val(JSON.stringify(this.data));
        },

        preview: function($target, emulateImageData, emulateCropData) {
            let $canvas = $target.parent(),
                $container = $canvas.parent(),
                containerWidth = $container.width(),
                containerHeight = $container.height(),
                imageData = emulateImageData || this.imageData,
                cropData = $.extend({}, emulateCropData || this.cropData), // Clone one to avoid changing it
                aspectRatio = cropData.width / cropData.height,
                canvasWidth = containerWidth,
                scaledRatio;

            if (canvasWidth == 0 || imageData.naturalWidth == 0 || imageData.naturalHeight == 0) {
                return;
            }

            if (containerHeight * aspectRatio <= containerWidth) {
                canvasWidth = containerHeight * aspectRatio;
            }

            scaledRatio = cropData.width / canvasWidth;

            $target.css({
                maxWidth: imageData.naturalWidth / scaledRatio,
                maxHeight: imageData.naturalHeight / scaledRatio
            });

            this.center($target);
        },

        center: function($target, reset) {
            $target.each(function() {
                let $this = $(this),
                    $canvas = $this.parent(),
                    $container = $canvas.parent();

                function center() {
                    let containerHeight = $container.height(),
                        canvasHeight = $canvas.height(),
                        marginTop = 'auto';

                    if (canvasHeight < containerHeight) {
                        marginTop = (containerHeight - canvasHeight) / 2;
                    }

                    $canvas.css('margin-top', marginTop);
                }

                if (reset) {
                    $canvas.add($this).removeAttr('style');
                }

                if (this.complete) {
                    center.call(this);
                } else {
                    this.onload = center;
                }
            });
        },

        autoCrop: function(url, data) {
            let cropData = this.cropData,
                cropOptions = this.data[this.options.key],
                _this = this;

            this.$list
                .find('img')
                .not(this.$target)
                .each(function() {
                    let $this = $(this),
                        sizeName = $this.data('sizeName');

                    if ($.inArray(sizeName, data) > -1) {
                        cropOptions[sizeName] = $.extend({}, cropData);

                        if (url) {
                            _this.center($this.attr('src', url), true);
                        } else {
                            _this.preview($this);
                        }
                    }
                });
        },

        destroy: function() {
            if (!isSVG) {
                this.unbind();
                this.unbuild();
            }
            this.$element.removeData(NAMESPACE);
        }
    };

    QorCropper.DEFAULTS = {
        parent: false,
        output: false,
        list: false,
        key: 'data',
        data: null,
        text: {
            title: 'Crop the image',
            ok: 'OK',
            cancel: 'Cancel'
        }
    };

    QorCropper.TOGGLE = `<div class="qor-cropper__toggle">
            <div class="qor-cropper__toggle--crop"><i class="material-icons">crop</i></div>
            <div class="qor-cropper__toggle--delete"><i class="material-icons">delete</i></div>
        </div>`;

    QorCropper.ALERT = `<div class="qor-fieldset__alert">
            <button class="mdl-button mdl-button--accent qor-fieldset__undo" type="button">Undo delete</button>
        </div>`;

    QorCropper.CANVAS = '<div class="qor-cropper__canvas"></div>';
    QorCropper.LIST = '<ul><li><img></li></ul>';
    QorCropper.FILE_LIST = `<div class="qor-file__list-item">
                                <span><span>{{filename}}</span></span>
                                <div class="qor-cropper__toggle">
                                    <div class="qor-cropper__toggle--delete"><i class="material-icons">delete</i></div>
                                </div>
                            </div>`;
    QorCropper.MODAL = `<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">
            <div class="mdl-card mdl-shadow--2dp" role="document">
                <div class="mdl-card__title">
                    <h2 class="mdl-card__title-text">$[title]</h2>
                </div>
                <div class="mdl-card__supporting-text">
                    <div class="qor-cropper__wrapper"></div>
                    <div class="qor-cropper__options">
                        <p>Sync cropping result to: <label><input type="checkbox" class="qor-cropper__options-toggle" checked/> All</label></p>
                    </div>
                </div>
                <div class="mdl-card__actions mdl-card--border">
                    <a class="mdl-button mdl-button--colored mdl-button--raised qor-cropper__save">$[ok]</a>
                    <a class="mdl-button mdl-button--colored" data-dismiss="modal">$[cancel]</a>
                </div>
                <div class="mdl-card__menu">
                    <button class="mdl-button mdl-button--icon" data-dismiss="modal" aria-label="close">
                        <i class="material-icons">close</i>
                    </button>
                </div>
            </div>
        </div>`;

    QorCropper.plugin = function(option) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                options,
                fn;

            if (!data) {
                if (!$.fn.cropper) {
                    return;
                }

                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorCropper(this, options)));
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        let selector = '.qor-file__input',
            options = {
                parent: '.qor-file',
                output: '.qor-file__options',
                list: '.qor-file__list',
                key: 'CropOptions'
            };

        $(document)
            .on(EVENT_ENABLE, function(e) {
                QorCropper.plugin.call($(selector, e.target), options);
            })
            .on(EVENT_DISABLE, function(e) {
                QorCropper.plugin.call($(selector, e.target), 'destroy');
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorCropper;
});
