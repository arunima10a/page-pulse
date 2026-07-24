$(function () {
    const $btn = $('#auditBtn');
    const $input = $('#targetUrl');
    const $error = $('#errorMessage');
    const $results = $('#results');
    const $led = $('#ledIndicator');
    const $statusText = $('#statusText');
    const $ecgStrip = $('#ecgStrip');
  
    function setLed(state) {
      $led.removeClass('standby scanning ok alert').addClass(state);
    }
  
    function setEcg(state) {
      $ecgStrip.removeClass('is-scanning is-alert');
      if (state === 'scanning') $ecgStrip.addClass('is-scanning');
      if (state === 'alert') $ecgStrip.addClass('is-alert');
    }
  
    function isValidUrl(value) {
      try {
        const u = new URL(value);
        return u.protocol === 'http:' || u.protocol === 'https:';
      } catch (e) {
        return false;
      }
    }
  
    // Animation of a number counting up to its final value
    function animateNumber($el, endVal, duration) {
      const startVal = 0;
      const startTime = performance.now();
      endVal = Number(endVal) || 0;
  
      function step(now) {
        const progress = Math.min((now - startTime) / duration, 1);
        const eased = 1 - Math.pow(1 - progress, 3); // ease-out cubic
        const current = Math.round(startVal + (endVal - startVal) * eased);
        $el.text(current.toLocaleString());
        if (progress < 1) requestAnimationFrame(step);
      }
      requestAnimationFrame(step);
    }
  
    $btn.on('click', runAudit);
    $input.on('keydown', function (e) {
      if (e.key === 'Enter') runAudit();
    });
  
    function runAudit() {
      const url = $input.val().trim();
  
      if ($btn.prop('disabled')) return;
  
      if (!url) {
        $input.trigger('focus');
        return;
      }
  
      if (!isValidUrl(url)) {
        showError('Enter a full URL, including https://');
        return;
      }
  
    
      $error.addClass('hidden');
      $results.addClass('hidden');
      $btn.prop('disabled', true).addClass('is-loading');
      $btn.find('.btn-label').text('Scanning…');
      setLed('scanning');
      setEcg('scanning');
      $statusText.text('Reading vitals for ' + url + '…');
  
      $.ajax({
        url: '/api/audit?url=' + encodeURIComponent(url),
        method: 'GET',
        dataType: 'json'
      })
        .done(function (data) {
          renderReport(data, url);
        })
        .fail(function (xhr) {
          const msg = (xhr.responseJSON && xhr.responseJSON.error) || 'An unexpected error occurred.';
          showError(msg);
        })
        .always(function () {
          $btn.prop('disabled', false).removeClass('is-loading');
          $btn.find('.btn-label').text('Run scan');
        });
    }
  
    function showError(message) {
      $error.text(message).removeClass('hidden');
      $results.addClass('hidden');
      setLed('alert');
      setEcg('alert');
      $statusText.text('Scan failed');
    }
  
    function renderReport(data, requestedUrl) {
      const status = data.status || 0;
      const healthy = status >= 200 && status < 400;
  
      setLed(healthy ? 'ok' : 'alert');
      setEcg(healthy ? 'idle' : 'alert');
      $statusText.text((healthy ? 'Vitals stable — last scan: ' : 'Vitals abnormal — last scan: ') + requestedUrl);
  
      // Status + response time
      const $status = $('#resStatus');
      $status.text(status || '—').removeClass('ok alert').addClass(healthy ? 'ok' : 'alert');
  
      const $miniEcg = $('#miniEcg');
      $miniEcg.removeClass('ok alert').addClass(healthy ? 'ok' : 'alert');
  
      const ms = Math.round((data.response_time_ms || 0) / 1000000);
      animateNumber($('#resTime'), ms, 500);
  
      // Structure / accessibility / content — count up
      const h1 = data.h1_count || 0;
      const altMissing = data.images_missing_alt || 0;
      const words = data.word_count || 0;
  
      animateNumber($('#resH1'), h1, 500);
      animateNumber($('#resAlt'), altMissing, 500);
      animateNumber($('#resWords'), words, 700);
  
      $('#resH1Plural').text(h1 === 1 ? '' : 's');
      $('#resAltPlural').text(altMissing === 1 ? '' : 's');
  
      $('#resAlt')
        .removeClass('ok alert warn')
        .addClass(altMissing === 0 ? 'ok' : 'alert');
  
      // Metadata readout
      const $title = $('#resTitle');
      const $desc = $('#resDesc');
  
      if (data.title) {
        $title.text(data.title).removeClass('empty');
      } else {
        $title.text('No title tag found on this page.').addClass('empty');
      }
  
      if (data.meta_description) {
        $desc.text(data.meta_description).removeClass('empty');
      } else {
        $desc.text('No meta description found.').addClass('empty');
      }
  
      $results.removeClass('hidden');
    }
  });