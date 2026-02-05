export default {
  async fetch(request) {
    const url = new URL(request.url);
    const path = url.pathname.substring(1);

    let bytes = 100000000; // Default 100MB
    if (path) {
      const match = path.match(/^(\d+)([a-z]{0,2})$/i);
      if (!match) {
        return new Response("路径格式不正确", { status: 400 });
      }

      const size = parseInt(match[1], 10);
      const unit = match[2].toLowerCase();
      const multipliers = {
        'k': 1000, 'kb': 1000,
        'm': 1000000, 'mb': 1000000,
        'g': 1000000000, 'gb': 1000000000
      };
      bytes = size * (multipliers[unit] || 1);
    }

    const targetUrl = `https://speed.cloudflare.com/__down?bytes=${bytes}`;
    const headers = new Headers(request.headers);
    headers.set('referer', 'https://speed.cloudflare.com/');

    return fetch(new Request(targetUrl, {
      method: request.method,
      headers: headers,
      body: request.body
    }));
  }
};


