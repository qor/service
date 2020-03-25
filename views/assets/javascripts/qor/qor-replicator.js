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

    let _ = window._,
        NAMESPACE = 'qor.replicator',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_SUBMIT = 'submit.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SLIDEOUTBEFORESEND = 'slideoutBeforeSend.qor.slideout.replicator',
        EVENT_SELECTCOREBEFORESEND = 'selectcoreBeforeSend.qor.selectcore.replicator bottomsheetBeforeSend.qor.bottomsheets.replicator',
        EVENT_REPLICATOR_ADDED = 'added.' + NAMESPACE,
        EVENT_REPLICATORS_ADDED = 'addedMultiple.' + NAMESPACE,
        EVENT_REPLICATORS_ADDED_DONE = 'addedMultipleDone.' + NAMESPACE,
        CLASS_CONTAINER = '.qor-fieldset-container';

    function QorReplicator(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorReplicator.DEFAULTS, $.isPlainObject(options) && options);
        this.index = 0;
        this.init();
    }

    QorReplicator.prototype = {
        constructor: QorReplicator,

        init: function() {
            let $element = this.$element,
                $template = $element.find('> .qor-field__block > .qor-fieldset--new'),
                fieldsetName;

            this.singlePage = !($element.closest('.qor-slideout').length && $element.closest('.qor-bottomsheets').length);
            this.maxitems = $element.data('maxItem');
            this.isSortable = $element.hasClass('qor-fieldset-sortable');

            if (!$template.length || $element.closest('.qor-fieldset--new').length) {
                return;
            }

            // Should destroy all components here
            $template.trigger('disable');
            // remove data-select2-id attribute or select2 will disable all previous instance
            $template.find('select[data-toggle]').removeAttr('data-select2-id');

            // if have isMultiple data value or template length large than 1
            this.isMultipleTemplate = $element.data('isMultiple');

            if (this.isMultipleTemplate) {
                this.fieldsetName = [];
                this.template = {};
                this.index = [];

                $template.each((i, ele) => {
                    fieldsetName = $(ele).data('fieldsetName');
                    if (fieldsetName) {
                        this.template[fieldsetName] = $(ele).prop('outerHTML');
                        this.fieldsetName.push(fieldsetName);
                    }
                });

                this.parseMultiple();
            } else {
                this.parse($template.prop('outerHTML'));
            }

            $template.hide();
            this.bind();
            this.resetButton();
            this.resetPositionButton();
        },

        resetPositionButton: function() {
            let sortableButton = this.$element.find('> .qor-sortable__button');

            if (this.isSortable) {
                if (this.getCurrentItems() > 1) {
                    sortableButton.show();
                } else {
                    sortableButton.hide();
                }
            }
        },

        getCurrentItems: function() {
            return this.$element.find('> .qor-field__block > .qor-fieldset').not('.qor-fieldset--new,.is-deleted').length;
        },

        toggleButton: function(isHide) {
            let $button = this.$element.find('> .qor-field__block > .qor-fieldset__add');

            if (isHide) {
                $button.hide();
            } else {
                $button.show();
            }
        },

        resetButton: function() {
            if (this.maxitems <= this.getCurrentItems()) {
                this.toggleButton(true);
            } else {
                this.toggleButton();
            }
        },

        parse: function($tmp) {
            let template;

            if (!$tmp) {
                return;
            }
            template = this.initTemplate($tmp);

            this.template = template.template;
            this.index = template.index;
        },

        parseMultiple: function() {
            let template,
                name,
                fieldsetName = this.fieldsetName;

            for (let i = 0, len = fieldsetName.length; i < len; i++) {
                name = fieldsetName[i];
                template = this.initTemplate(this.template[name]);
                this.template[name] = template.template;
                this.index.push(template.index);
            }

            this.multipleIndex = _.max(this.index);
        },

        initTemplate: function(template) {
            let i,
                deepLevel = this.$element.parents(CLASS_CONTAINER).length;

            template = template.replace(/(\w+)\="(\S*\[\d+\]\S*)"/g, function(attribute, name, value) {
                value = value.replace(/^(\S*)\[(\d+)\]([^\[\]]*)$/, function(input, prefix, index) {
                    if (input === value) {
                        if (name === 'name' && !i) {
                            i = index;
                        }

                        if (deepLevel) {
                            // assume input = QorResource.SerializableMeta.Menus[1].SubMenus[2].Items[3].URL
                            // if deepLevel = 1, input should be QorResource.SerializableMeta.Menus[1].SubMenus[{{index}}].Items[3].URL
                            // if deepLevel = 2, input should be QorResource.SerializableMeta.Menus[1].SubMenus[2].Items[{{index}}].URL

                            let newInput = '',
                                splitStr = input.split(/\[\d+\]/), // ["QorResource.SerializableMeta.Menus", ".SubMenus", ".Items", ".URL"]
                                sortNumbers = input.match(/\[\d+\]/g); // ["[1]", "[2]", "[3]"]

                            for (let j = 0; j < splitStr.length; j++) {
                                let str = '';
                                if (j === deepLevel) {
                                    str = '[{{index}}]';
                                } else if (j < sortNumbers.length) {
                                    str = sortNumbers[j];
                                }
                                newInput += splitStr[j] + str;
                            }

                            return newInput;
                        } else {
                            return input.replace(/\[\d+\]/, '[{{index}}]');
                        }
                    }
                });

                return name + '="' + value + '"';
            });

            return {
                template: template,
                index: parseFloat(i) + 5 //make sure the index is different from original.
            };
        },

        bind: function() {
            let options = this.options;

            this.$element.on(EVENT_CLICK, options.addClass, $.proxy(this.add, this)).on(EVENT_CLICK, options.delClass, $.proxy(this.del, this));

            this.singlePage && $(document).on(EVENT_SUBMIT, '.mdl-layout__container form', this.clearFieldData);
            $(document)
                .on(EVENT_SLIDEOUTBEFORESEND, '.qor-slideout', this.clearFieldDataInSlideout)
                .on(EVENT_SELECTCOREBEFORESEND, this.clearFieldDataInBottomsheet);
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK);

            this.singlePage && $(document).off(EVENT_SUBMIT, '.mdl-layout__container form', this.clearFieldData);
            $(document)
                .off(EVENT_SLIDEOUTBEFORESEND, '.qor-slideout', this.clearFieldDataInSlideout)
                .off(EVENT_SELECTCOREBEFORESEND, this.clearFieldDataInBottomsheet);
        },

        clearFieldData: function() {
            $('.qor-fieldset--new').remove();
        },

        clearFieldDataInSlideout: function() {
            $('.qor-slideout .qor-fieldset--new').remove();
        },

        clearFieldDataInBottomsheet: function() {
            $('.qor-bottomsheets .qor-fieldset--new').remove();
        },

        add: function(e, data, isAutomatically) {
            let options = this.options,
                $item,
                template,
                $target = $(e.target).closest(options.addClass);

            if (this.maxitems <= this.getCurrentItems()) {
                return false;
            }

            if (this.isMultipleTemplate) {
                let templateName = $target.data('template'),
                    parents = $target.closest(this.$element),
                    parentsChildren = parents.children(options.childrenClass),
                    $fieldset = $target.closest(options.childrenClass).children('fieldset');

                template = this.template[templateName];
                $item = $(template.replace(/\{\{index\}\}/g, this.multipleIndex));

                // get input kind from add button then add into QorResource.Rules[1].Kind input
                for (let dataKey in $target.data()) {
                    if (dataKey.match(/^sync/)) {
                        let k = dataKey.replace(/^sync/, '');
                        $item.find("input[name*='." + k + "']").val($target.data(dataKey));
                    }
                }

                if ($fieldset.length) {
                    $fieldset.last().after($item.show());
                } else {
                    parentsChildren.prepend($item.show());
                }
                $item.data('itemIndex', this.multipleIndex).removeClass('qor-fieldset--new');
                this.multipleIndex++;
            } else {
                if (!isAutomatically) {
                    $item = this.addSingle();
                    $target.before($item.show());
                    this.index++;
                } else {
                    if (data && data.length) {
                        this.addMultiple(data);
                        $(document).trigger(EVENT_REPLICATORS_ADDED_DONE);
                    }
                }
            }

            if (!isAutomatically) {
                $item.trigger('enable');
                $(document).trigger(EVENT_REPLICATOR_ADDED, [$item]);
                e.stopPropagation();
            }

            this.resetPositionButton();
            this.resetButton();
        },

        addMultiple: function(data) {
            let $item;

            for (let i = 0, len = data.length; i < len; i++) {
                $item = this.addSingle();
                this.index++;
                $(document).trigger(EVENT_REPLICATORS_ADDED, [$item, data[i]]);
            }
        },

        addSingle: function() {
            let $item,
                $element = this.$element;

            if (!this.template) {
                return;
            }

            $item = $(this.template.replace(/\{\{index\}\}/g, this.index));
            // add order property for sortable fieldset
            if (this.isSortable) {
                let order = $element.find('> .qor-field__block > .qor-sortable__item').not('.qor-fieldset--new').length;
                $item
                    .attr('order-index', order)
                    .attr('order-item', `item_${order}`)
                    .css('order', order);
            }

            $item.data('itemIndex', this.index).removeClass('qor-fieldset--new');

            return $item;
        },

        del: function(e) {
            let options = this.options,
                $item = $(e.target).closest(options.itemClass),
                $alert,
                that = this,
                message = {
                    confirm:
                        $(e.target)
                            .closest(options.delClass)
                            .data('confirm') || 'Are you sure?'
                };

            QOR.qorConfirm(message, function(confirm) {
                if (confirm) {
                    $item
                        .addClass('is-deleted')
                        .children(':visible')
                        .addClass('hidden')
                        .hide();
                    $alert = $(options.alertTemplate.replace('{{name}}', that.parseName($item)));
                    $alert.find(options.undoClass).one(
                        EVENT_CLICK,
                        function() {
                            if (that.maxitems <= that.getCurrentItems()) {
                                window.QOR.qorConfirm(that.$element.data('maxItemHint'));
                                return false;
                            }

                            $item.find('> .qor-fieldset__alert').remove();
                            $item
                                .removeClass('is-deleted')
                                .children('.hidden')
                                .removeClass('hidden')
                                .show();
                            that.resetButton();
                            that.resetPositionButton();
                        }.bind(this)
                    );
                    that.resetButton();
                    that.resetPositionButton();
                    $item.append($alert);
                }
            });
        },

        parseName: function($item) {
            let name = $item.find('input[name]').attr('name') || $item.find('textarea[name]').attr('name');

            if (name) {
                return name.replace(/[^\[\]]+$/, '');
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorReplicator.DEFAULTS = {
        itemClass: '.qor-fieldset',
        newClass: '.qor-fieldset--new',
        addClass: '.qor-fieldset__add',
        delClass: '.qor-fieldset__delete',
        childrenClass: '.qor-field__block',
        undoClass: '.qor-fieldset__undo',
        alertTemplate:
            '<div class="qor-fieldset__alert">' +
            '<input type="hidden" name="{{name}}._destroy" value="1">' +
            '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button">Undo delete</button>' +
            '</div>'
    };

    QorReplicator.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.call(data);
            }
        });
    };

    $(function() {
        let selector = CLASS_CONTAINER;
        let options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorReplicator.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorReplicator.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorReplicator;
});
