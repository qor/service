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

    let $window = $(window),
        NAMESPACE = 'qor.fixer',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_RESIZE = 'resize.' + NAMESPACE,
        EVENT_SCROLL = 'scroll.' + NAMESPACE,
        CLASS_FIXED_TABLE = 'qor-table-fixed-header',
        CLASS_HEADER = '.qor-page__header';

    function QorFixer(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorFixer.prototype = {
        constructor: QorFixer,

        init: function() {
            var options = this.options;
            var $this = this.$element;
            if (this.isNeedBuild()) {
                return;
            }
            this.$thead = $this.find('> thead');
            this.$tbody = $this.find('> tbody');

            this.$header = $(options.header);
            this.$subHeader = $(options.subHeader);
            this.$content = $(options.content);
            this.marginBottomPX = parseInt(this.$subHeader.css('marginBottom'));
            this.paddingHeight = options.paddingHeight;

            this.resize();
            this.bind();
        },

        bind: function() {
            this.$content.on(EVENT_SCROLL, this.toggle.bind(this));
            $window.on(EVENT_RESIZE, this.resize.bind(this));
        },

        unbind: function() {
            this.$content.off(EVENT_SCROLL, this.toggle).off(EVENT_RESIZE, this.resize);
        },

        isNeedBuild: function() {
            var $this = this.$element;
            // disable fixer if have multiple tables or in search page or in media library list page
            if (
                $('.qor-page__body .qor-js-table').length > 1 ||
                $('.qor-global-search--container').length > 0 ||
                $this.hasClass('qor-table--medialibrary') ||
                $this.is(':hidden') ||
                $this.find('tbody > tr:visible').length <= 1 ||
                $this.data("disable-fixer")
            ) {
                return true;
            }
            return false;
        },

        build: function() {
            let headerWidth = [],
                $items = this.$tbody.find('> tr:first').children();

            $items.each(function() {
                let tdWidth = $(this).outerWidth();
                $(this).outerWidth(tdWidth);
                headerWidth.push(tdWidth);
            });

            this.$thead
                .find('>tr')
                .children()
                .each(function(i) {
                    $(this).outerWidth(headerWidth[i]);
                });
        },

        toggle: function() {
            if (!this.$content.length) {
                return;
            }
            let $element = this.$element,
                $thead = this.$thead,
                scrollTop = this.$content.scrollTop(),
                offsetTop = this.$subHeader.outerHeight() + this.paddingHeight + this.marginBottomPX,
                headerHeight = $('.qor-page__header').outerHeight(),
                pageTop = this.$content.offset().top + $(CLASS_HEADER).height();

            if (scrollTop > offsetTop - headerHeight) {
                $thead.css({top: pageTop});
                $element.addClass(CLASS_FIXED_TABLE);
            } else {
                $thead.css({top: 0});
                $element.removeClass(CLASS_FIXED_TABLE);
            }
        },

        resize: function() {
            this.build();
            this.toggle();
        },

        destroy: function() {
            if (this.buildCheck()) {
                return;
            }
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorFixer.DEFAULTS = {
        header: false,
        content: false
    };

    QorFixer.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorFixer(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.call(data);
            }
        });
    };

    $(function() {
        var selector = '.qor-js-table';
        var options = {
            header: '.mdl-layout__header',
            subHeader: '.qor-page__header',
            content: '.mdl-layout__content',
            paddingHeight: 2 // Fix sub header height bug
        };

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorFixer.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorFixer.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorFixer;
});
