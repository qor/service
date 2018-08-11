$(function () {
  $('.qor-mobile--show-actions').on('click', function () {
    $('.qor-page__header').toggleClass('actions-show');
  });

  $("body").on("focus",".js-custom_datepicker",function(){
    if(!$(this).data("is_init")){
      $(this).datetimepicker({format:'Y-m-d H:i:00',timepicker:true});
      $.datetimepicker.setLocale('ch');
      $(this).data("is_init", true);
      $(this).focus();
    }
  });

  $("body").on("focus",".js-date-picker",function(){
    if(!$(this).data("is_init")){
      $(this).datetimepicker({format:'Y-m-d',timepicker:false});
      $.datetimepicker.setLocale('ch');
      $(this).data("is_init", true);
      $(this).focus();
    }
  });
});
