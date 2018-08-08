$(function () {
  $('.qor-mobile--show-actions').on('click', function () {
    $('.qor-page__header').toggleClass('actions-show');
  });

  if($(".js-custom_datepicker").get(0)) {
    $(".js-custom_datepicker").datetimepicker({ format: 'Y-m-d H:i', timepicker: true});
    $.datetimepicker.setLocale('ch');
  }
});
