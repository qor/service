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

    var NAMESPACE = 'qor.chooser';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorChooser(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorChooser.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorChooser.prototype = {
        constructor: QorChooser,

        init: function() {
            let $this = this.$element,
                select2Data = $this.data(),
                resetSelect2Width,
                option = {
                    minimumResultsForSearch: 8,
                    dropdownParent: $this.parent()
                };
            let getSelect2AjaxDynamicURL = window.getSelect2AjaxDynamicURL;
            let remoteImage = select2Data.remoteImage;

            if (select2Data.remoteData) {
                option.ajax = $.fn.select2.ajaxCommonOptions(select2Data);
                if (getSelect2AjaxDynamicURL && $.isFunction(getSelect2AjaxDynamicURL)) {
                    option.ajax.url = function() {
                        return getSelect2AjaxDynamicURL(select2Data);
                    };
                } else {
                    option.ajax.url = select2Data.remoteUrl;
                }

                option.templateResult = function(data) {
                    let tmpl = $this.parents('.qor-field').find('[name="select2-result-template"]');
                    return $.fn.select2.ajaxFormatResult(data, tmpl, remoteImage);
                };

                option.templateSelection = function(data) {
                    if (data.loading) return data.text;
                    let tmpl = $this.parents('.qor-field').find('[name="select2-selection-template"]');
                    return $.fn.select2.ajaxFormatResult(data, tmpl, remoteImage);
                };
            }

            $this
                .on('select2:select', function(evt) {
                    $(evt.target).attr('chooser-selected', 'true');
                })
                .on('select2:unselect', function(evt) {
                    $(evt.target).attr('chooser-selected', '');
                });

            $this.select2(option);

            if ($this.is('[selectall]')) {
                if(select2Data.remoteUrl) {
                    $.ajax({
                        url: select2Data.remoteUrl,
                        dataType: 'json'
                    }).then(function (data) {
                        data.forEach(function(item) {  
                            if($this.find('option[value="'+item.ID+'"]').length < 1) {
                                $this.append(new Option(item.Name, item.ID, false, false));
                            }
                        });
                    });
                }

                $this.next('.select2-container').find('.select2-selection--multiple')
                    .prepend('<span class="select2-selection__select-all" title="Select all items">&equiv;</span>')
                    .find('.select2-selection__select-all').on('click', function(evt) {
                        var $select = $(evt.target).closest('.select2-container').prev('select[data-toggle="'+NAMESPACE+'"]');
                        evt.stopPropagation();
                        $select.find('option').prop('selected', 'selected');
                        $select.trigger('change');
                    });
            }

            // reset select2 container width
            this.resetSelect2Width();
            resetSelect2Width = window._.debounce(this.resetSelect2Width.bind(this), 300);
            $(window).resize(resetSelect2Width);

            if ($this.val()) {
                $this.attr('chooser-selected', 'true');
            }
        },

        resetSelect2Width: function() {
            var $container,
                select2 = this.$element.data().select2;
            if (select2 && select2.$container) {
                $container = select2.$container;
                $container.width($container.parent().width());
            }
        },

        destroy: function() {
            this.$element.select2('destroy').removeData(NAMESPACE);
        }
    };

    QorChooser.DEFAULTS = {};

    QorChooser.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorChooser(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = 'select[data-toggle="qor.chooser"]';

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorChooser.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorChooser.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorChooser;
});
