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
    let Mustache = window.Mustache,
        QOR = window.QOR,
        NAMESPACE = 'qor.action',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_UNDO = 'undo.' + NAMESPACE,
        CLASS_ACTION_FORMS = '.qor-action-forms',
        CLASS_MENU_ACTIONS = '[data-ajax-form="true"][data-method]',
        CLASS_BUTTON_BULKS = '.qor-action-bulk-buttons',
        CLASS_TABLE = '.qor-page .qor-table-container',
        CLASS_TABLE_BULK = '.qor-table--bulking',
        CLASS_TABLE_BULK_TR = '.qor-table--bulking tbody tr',
        CLASS_IS_UNDO = 'is_undo',
        CLASS_TABLE_MDL = 'mdl-data-table--selectable',
        CLASS_SLIDEOUT = '.qor-slideout',
        ACTION_FORM_DATA = 'primary_values[]',
        CLASS_HEADER_TOGGLE = '.qor-page__header .qor-actions, .qor-page__header .qor-search-container';

    function QorAction(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorAction.DEFAULTS, $.isPlainObject(options) && options);
        this.ajaxForm = {};
        this.init();
    }

    QorAction.prototype = {
        constructor: QorAction,

        init: function() {
            this.bind();
            this.initActions();
        },

        bind: function() {
            this.$element.on(EVENT_CLICK, '.qor-action--bulk', this.renderBulkTable.bind(this)).on(EVENT_CLICK, '.qor-action--exit-bulk', this.removeBulkTable.bind(this));

            $(document)
                .on(EVENT_CLICK, CLASS_TABLE_BULK_TR, this.handleBulkTableClick.bind(this))
                .on(EVENT_CLICK, CLASS_MENU_ACTIONS, this.clickAjaxButton.bind(this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK);

            $(document)
                .off(EVENT_CLICK, CLASS_TABLE_BULK_TR, this.handleBulkTableClick)
                .off(EVENT_CLICK, CLASS_MENU_ACTIONS, this.clickAjaxButton);
        },

        initActions: function() {
            if (!$(CLASS_TABLE).find('table').length) {
                $(CLASS_BUTTON_BULKS).hide();
                $('.qor-page__header a.qor-action--button').hide();
            }
        },

        collectFormData: function() {
            let checkedInputs = $(CLASS_TABLE_BULK).find('.mdl-checkbox__input:checked'),
                formData = [],
                normalFormData = [],
                tempObj;

            if (checkedInputs.length) {
                checkedInputs.each(function() {
                    let id = $(this)
                        .closest('tr')
                        .data('primary-key');

                    tempObj = {};
                    if (id) {
                        formData.push({
                            name: ACTION_FORM_DATA,
                            value: id.toString()
                        });

                        tempObj[ACTION_FORM_DATA] = id.toString();
                        normalFormData.push(tempObj);
                    }
                });
            }
            this.ajaxForm.formData = formData;
            this.ajaxForm.normalFormData = normalFormData;
            return this.ajaxForm;
        },

        actionSubmit: function($action) {
            this.submit($action);
            return false;
        },

        handleBulkTableClick: function(e) {
            let $target = $(e.target).closest('tr'),
                $firstTd = $target.find('td').first(),
                $checkbox = $firstTd.find('.mdl-js-checkbox');

            $checkbox.toggleClass('is-checked');
            $target.toggleClass('is-selected');
            $firstTd.find('input').prop('checked', $checkbox.hasClass('is-checked'));

            return false;
        },

        adjustPageBodyStyle: function(isRender) {
            let $pageHeader = $('.qor-page > .qor-page__header'),
                $pageBody = $('.qor-page > .qor-page__body'),
                triggerHeight = $pageHeader.find('.qor-page-subnav__header').length ? 96 : 48;

            if (isRender) {
                if ($pageHeader.height() > triggerHeight) {
                    $pageBody.css('padding-top', $pageHeader.height());
                }
            } else {
                if (parseInt($pageBody.css('padding-top')) > triggerHeight) {
                    $pageBody.css('padding-top', '');
                }
            }
        },

        renderBulkTable: function() {
            let $body = $('body');

            if ($body.hasClass('qor-slideout-open')) {
                $body.data('qor.slideout').hide();
            }

            $('.qor-table__inner-list').remove();
            this.toggleBulkButtons();
            this.enableTableMDL();
            this.adjustPageBodyStyle(true);
        },

        removeBulkTable: function() {
            this.toggleBulkButtons();
            this.disableTableMDL();
            this.adjustPageBodyStyle();
        },

        enableTableMDL: function() {
            $(CLASS_TABLE)
                .find('table')
                .removeAttr('data-upgraded')
                .addClass(CLASS_TABLE_MDL)
                .trigger('enable');
        },

        disableTableMDL: function() {
            $(CLASS_TABLE)
                .find('table')
                .removeClass(CLASS_TABLE_MDL)
                .find('tr')
                .removeClass('is-selected')
                .find('td:first,th:first')
                .remove();
        },

        toggleBulkButtons: function() {
            this.$element.find(CLASS_ACTION_FORMS).toggle();
            $(CLASS_BUTTON_BULKS)
                .find('button')
                .toggleClass('hidden');

            $(CLASS_TABLE)
                .toggleClass('qor-table--bulking')
                .find('.qor-table__actions')
                .toggle();

            $(CLASS_HEADER_TOGGLE).toggle();
        },

        clickAjaxButton: function(e) {
            let $target = $(e.target);

            this.collectFormData();
            this.ajaxForm.properties = $target.data();
            this.submit($target);
            return false;
        },

        renderFlashMessage: function(data) {
            let flashMessageTmpl = QorAction.FLASHMESSAGETMPL;
            Mustache.parse(flashMessageTmpl);
            return Mustache.render(flashMessageTmpl, data);
        },

        submit: function($actionButton) {
            let _this = this,
                ajaxForm = this.ajaxForm || {},
                properties = ajaxForm.properties || $actionButton.data();

            if (properties.fromIndex && (!ajaxForm.formData || !ajaxForm.formData.length)) {
                QOR.qorConfirm(ajaxForm.properties.errorNoItem);
                return;
            }

            if (properties.confirm) {
                QOR.qorConfirm(properties, function(confirm) {
                    if (confirm) {
                        _this.handleAjaxSubmit(ajaxForm, $actionButton);
                    } else {
                        return;
                    }
                });
            } else {
                this.handleAjaxSubmit(ajaxForm, $actionButton);
            }
        },

        handleAjaxSubmit: function(ajaxForm, $actionButton) {
            let _this = this,
                $element = this.$element,
                properties = ajaxForm.properties || $actionButton.data(),
                url = properties.url,
                undoUrl = properties.undoUrl,
                isUndo = $actionButton.hasClass(CLASS_IS_UNDO),
                isInSlideout = $actionButton.closest(CLASS_SLIDEOUT).length,
                needDisableButtons = $element.length && !isInSlideout;
            if (isUndo) {
                url = undoUrl; // notification has undo url
            }
            if (isInSlideout==0){
                isInSlideout = $element.length;
                needDisableButtons = $element.length && !isInSlideout;
            }

            $.ajax(url, {
                method: properties.method,
                data: ajaxForm.formData,
                dataType: properties.datatype || 'json',
                beforeSend: function() {
                    if (undoUrl) {
                        $actionButton.prop('disabled', true);
                    } else if (needDisableButtons) {
                        _this.switchButtons($element, 1);
                    }
                },
                success: function(data) {
                    // has undo action
                    if (undoUrl) {
                        $element.trigger(EVENT_UNDO, [$actionButton, isUndo, data]);
                        isUndo ? $actionButton.removeClass(CLASS_IS_UNDO) : $actionButton.addClass(CLASS_IS_UNDO);
                        $actionButton.prop('disabled', false);
                        return;
                    }

                    window.location.reload();
                },
                error: function(err) {
                    if (err.status == 200) {
                        return;
                    }
                    if (undoUrl) {
                        $actionButton.prop('disabled', false);
                    } else if (needDisableButtons) {
                        _this.switchButtons($element);
                    }

                    QOR.handleAjaxError(err);
                },
                complete: function(response) {
                    let contentType = response.getResponseHeader('content-type'),
                        disposition = response.getResponseHeader('Content-Disposition');

                    // handle file download from form submit
                    if (disposition && disposition.indexOf('attachment') !== -1) {
                        var fileNameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/,
                            matches = fileNameRegex.exec(disposition),
                            fileData = {},
                            fileName = '';

                        if (matches != null && matches[1]) {
                            fileName = matches[1].replace(/['"]/g, '');
                        }

                        if (properties.method) {
                            fileData = $.extend({}, ajaxForm.normalFormData, {
                                _method: properties.method
                            });
                        }

                        QOR.qorAjaxHandleFile(url, contentType, fileName, fileData);

                        if (undoUrl) {
                            $actionButton.prop('disabled', false);
                        } else {
                            _this.switchButtons($element);
                        }
                    }
                }
            });
        },

        switchButtons: function($element, disbale) {
            let needDisbale = disbale ? true : false;
            $element.find('.qor-action-button').prop('disabled', needDisbale);
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorAction.DEFAULTS = {};

    $.fn.qorSliderAfterShow.qorInsertActionData = function(url, html) {
        let $action = $(html).find('[data-toggle="qor-action-slideout"]'),
            $actionForm = $action.find('form'),
            $checkedItem = $(CLASS_TABLE_BULK).find('.mdl-checkbox__input:checked');

        if ($action.length && $checkedItem.length) {
            // insert checked value into sliderout form
            $checkedItem.each(function() {
                let id = $(this)
                    .closest('tr')
                    .data('primary-key');

                if (id) {
                    $actionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + id + '" />');
                }
            });
        }
    };

    QorAction.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorAction(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.call(data);
            }
        });
    };

    $(function() {
        let options = {},
            selector = '[data-toggle="qor.action.bulk"]';

        if (!$(selector).length) {
            $(document).on(EVENT_CLICK, CLASS_MENU_ACTIONS, function(e) {
                new QorAction().actionSubmit($(e.target));
                return false;
            });
        }

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorAction.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorAction.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorAction;
});
