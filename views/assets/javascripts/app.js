$(function(){"use strict";$(document).on("click.qor.alert",'[data-dismiss="alert"]',function(){$(this).closest(".qor-alert").remove()}),setTimeout(function(){$('.qor-alert[data-dismissible="true"]').remove()},5e3)}),$(function(){"use strict";var e=$(".qor-page__body > .qor-form-container > form");$(".qor-error > li > label").each(function(){var a=$(this),o=a.attr("for");o&&e.find("#"+o).closest(".qor-field").addClass("is-error").append(a.clone().addClass("qor-field__error"))})}),$(function(){"use strict";var e='<div class="qor-dialog qor-dialog--global-search" tabindex="-1" role="dialog" aria-hidden="true"><div class="qor-dialog-content"><form action=[[actionUrl]]><div class="mdl-textfield mdl-js-textfield" id="global-search-textfield"><input class="mdl-textfield__input ignore-dirtyform" name="keyword" id="globalSearch" value="" type="text" placeholder="" /><label class="mdl-textfield__label" for="globalSearch">[[placeholder]]</label></div></form></div></div>';$(document).on("click",".qor-dialog--global-search",function(e){e.stopPropagation(),$(e.target).parents(".qor-dialog-content").size()||$(e.target).is(".qor-dialog-content")||$(".qor-dialog--global-search").remove()}),$(document).on("click",".qor-global-search--show",function(a){a.preventDefault();var o=$(this).data(),t=window.Mustache.render(e,o);$("body").append(t),window.componentHandler.upgradeElement(document.getElementById("global-search-textfield")),$("#globalSearch").focus()})}),$(function(){"use strict";$(".qor-menu-container").on("click","> ul > li > a",function(){var e=$(this),a=e.parent(),o=e.next("ul");o.length&&(o.hasClass("in")?(a.removeClass("is-expanded"),o.one("transitionend",function(){o.removeClass("collapsing in")}).addClass("collapsing").height(0)):(a.addClass("is-expanded"),o.one("transitionend",function(){o.removeClass("collapsing")}).addClass("collapsing in").height(o.prop("scrollHeight"))))}).find("> ul > li > a").each(function(){var e=$(this),a=e.parent(),o=e.next("ul");o.length&&(a.addClass("has-menu is-expanded"),o.addClass("collapse in").height(o.prop("scrollHeight")))}),$(".qor-page").find(".qor-page__header").size()&&($(".qor-page").addClass("has-header"),$("header.mdl-layout__header").addClass("has-action")),$(".qor-page .qor-page__header").height()>48&&$(".qor-page").css("padding-top",$(".qor-page .qor-page__header").height())}),$(function(){$(".qor-mobile--show-actions").on("click",function(){$(".qor-page__header").toggleClass("actions-show")})}),$(function(){"use strict";function e(){$("[data-url]").removeClass(s)}function a(e){$("[data-url]").removeClass(s),e.addClass(s)}var o,t,r=$("body"),s="is-selected",n=r.hasClass("qor-theme-slideout"),l=function(){return r.hasClass("qor-slideout-open")};r.qorBottomSheets(),n&&r.qorSlideout(),o=r.data("qor.slideout"),t=r.data("qor.bottomsheets"),$(document).on("click.qor.openUrl","[data-url]",function(r){var i=$(this),d=i.hasClass("qor-button--new"),c=i.hasClass("qor-button--edit"),h=i.is(".qor-table tr[data-url]"),u=i.hasClass("qor-action-button"),p=i.data();if(!($(r.target).hasClass("material-icons")||p.method&&"GET"!=p.method.toUpperCase())){if(h||d||c||"slideout"==p.openType){if(n){if(!i.hasClass(s))return o.open(p),a(i),!1;o.hide(),e()}else window.location=p("url");return}return l()||u||"bottom-sheet"==p.openType?(t.open(p),!1):n?(o.open(p),!1):(t.open(p),!1)}})}),$(function(){"use strict";var e=window.location;$(".qor-search").each(function(){var a=$(this),o=a.find(".qor-search__input"),t=a.find(".qor-search__clear"),r=!!o.val(),s=function(){var a=e.search.replace(new RegExp(o.attr("name")+"\\=?\\w*"),"");"?"==a?e.href=e.href.split("?")[0]:e.search=e.search.replace(new RegExp(o.attr("name")+"\\=?\\w*"),"")};a.closest(".qor-page__header").addClass("has-search"),$("header.mdl-layout__header").addClass("has-search"),t.on("click",function(){o.val()||r?s():a.removeClass("is-dirty")})})}),$(function(){"use strict";$(".qor-js-table .qor-table__content").each(function(){var e=$(this),a=e.width(),o=e.parent().width();a>=180&&a<o&&e.css("max-width",o)})});
//# sourceMappingURL=app.js.map
