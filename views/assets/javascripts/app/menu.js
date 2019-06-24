$(function() {
  'use strict';

  let menuDatas = [],
    storageName = 'qoradmin_menu_status',
    lastMenuStatus = localStorage.getItem(storageName);

  if (lastMenuStatus && lastMenuStatus.length) {
    menuDatas = lastMenuStatus.split(',');
  }

  $('.qor-menu-container')
    .on('click', '> ul > li a.has-submenu', function() {
      let $this = $(this),
        $li = $this.parent(),
        $ul = $this.next('ul'),
        menuName = $li.attr('qor-icon-name');

      if (!$ul.length) {
        return;
      }

      if ($ul.hasClass('in')) {
        menuDatas.push(menuName);
        $li.removeClass('is-expanded');

//        $ul
//          .one('transitionend', function() {
//            $ul.removeClass('collapsing in');
//          })
//          .addClass('collapsing')
//          .height(0);

        $ul.removeClass('collapsing in');

      } else {
        menuDatas = _.without(menuDatas, menuName);

        $li.addClass('is-expanded');
        //$ul
        //  .one('transitionend', function() {
        //    $ul.removeClass('collapsing');
        //  })
        //  .addClass('collapsing in')
        //  .height($ul.prop('scrollHeight'));

        $ul.addClass('collapsing in');

      }
      localStorage.setItem(storageName, menuDatas);
    })
    .find('> ul > li a.has-submenu')
    .each(function() {
      let $this = $(this),
        $li = $this.parent(),
        $ul = $this.next('ul'),
        menuName = $li.attr('qor-icon-name');

      if (!$ul.length) {
        return;
      }

      $ul.addClass('collapse');
      $li.addClass('has-menu');

      if (menuDatas.indexOf(menuName) != -1) {
        //$ul.height(0);
      } else {
        $li.addClass('is-expanded');
        //$ul.addClass('in').height($ul.prop('scrollHeight'));
        $ul.addClass('in');
      }
    });

  let $pageHeader = $('.qor-page > .qor-page__header'),
    $pageBody = $('.qor-page > .qor-page__body'),
    triggerHeight = $pageHeader.find('.qor-page-subnav__header').length ? 96 : 48;

  if ($pageHeader.length) {
    if ($pageHeader.height() > triggerHeight) {
      $pageBody.css('padding-top', $pageHeader.height());
    }

    $('.qor-page').addClass('has-header');
    $('header.mdl-layout__header').addClass('has-action');
  }
})

