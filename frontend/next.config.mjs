/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  distDir: 'dist',
  reactStrictMode: false,
  webpack: (config) => {
    // Add use-wallet dependency modules that cause "not found" errors
    config.externals.push('bufferutil', 'utf-8-validate', 'encoding');
    config.resolve.fallback = { fs: false };

    return config;
  },
};

export default nextConfig;
