'use strict';

const express = require('express');
const { isIPv6 } = require('net');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3000;

// Trust the first proxy (nginx) so req.ip reflects the real client IP
app.set('trust proxy', 1);

app.use(express.static(path.join(__dirname, 'public')));

app.get('/', (req, res) => {
  const clientIp = req.ip;
  // Strip IPv6-mapped IPv4 prefix (::ffff:1.2.3.4)
  const cleanIp = clientIp.replace(/^::ffff:/, '');
  const hasIpv6 = isIPv6(clientIp) && !clientIp.startsWith('::ffff:');

  res.send(buildPage(cleanIp, hasIpv6));
});

// JSON API endpoint for programmatic use
app.get('/api', (req, res) => {
  const clientIp = req.ip;
  const cleanIp = clientIp.replace(/^::ffff:/, '');
  const hasIpv6 = isIPv6(clientIp) && !clientIp.startsWith('::ffff:');

  res.json({
    hasIpv6,
    ip: cleanIp,
    type: hasIpv6 ? 'IPv6' : 'IPv4',
  });
});

function buildPage(ip, hasIpv6) {
  const verdict = hasIpv6 ? 'YES' : 'NO';
  const verdictClass = hasIpv6 ? 'verdict-yes' : 'verdict-no';

  const ipSection = hasIpv6
    ? `<p class="ip-label">Your IPv6 address</p>
       <p class="ip-value ipv6">${escapeHtml(ip)}</p>`
    : `<p class="ip-label">Your IPv4 address</p>
       <p class="ip-value ipv4">${escapeHtml(ip)}</p>
       <p class="no-explainer">Your device connected over IPv4, so either your ISP doesn't offer IPv6 yet, or it's disabled on your network.</p>`;

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Do I Have IPv6?</title>
  <link rel="stylesheet" href="/style.css">
</head>
<body>
  <main>
    <h1 class="site-title">Do I Have IPv6?</h1>

    <div class="verdict ${verdictClass}">${verdict}</div>

    <div class="ip-info">
      ${ipSection}
    </div>

    <div class="explainer">
      <p>IPv6 is the modern internet protocol that provides virtually unlimited addresses.
         Most major ISPs and networks now support it.</p>
      <p>You connected from <strong>${escapeHtml(ip)}</strong>, which is ${hasIpv6 ? 'an <strong>IPv6</strong> address' : 'an <strong>IPv4</strong> address'}.</p>
    </div>

    <footer>
      <a href="/api">JSON API</a> &mdash; <a href="https://github.com/jonwadsworth/doihaveipv6">source</a>
    </footer>
  </main>
</body>
</html>`;
}

function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

app.listen(PORT, () => {
  console.log(`doihaveipv6 running on port ${PORT}`);
});
