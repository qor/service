$(function () {
  $('.qor-mobile--show-actions').on('click', function () {
    $('.qor-page__header').toggleClass('actions-show');
  });

  $("body").on("focus",".js-custom_datepicker",function(){
    if(!$(this).data("is_init")){
      $(this).datetimepicker({format:'Y-m-d H:i',timepicker:true});
      $.datetimepicker.setLocale('ch');
      $(this).data("is_init", true);
      $(this).focus();
    }
  });
});
