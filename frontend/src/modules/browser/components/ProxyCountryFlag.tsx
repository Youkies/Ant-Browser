import * as FlagIcons from 'country-flag-icons/react/3x2'
import type { BrowserProxy, ProxyIPHealthResult } from '../types'

const PROXY_IP_HEALTH_CACHE_KEY = 'browser:proxyPool:ipHealthMap:v1'
const PROXY_IP_HEALTH_CACHE_TTL_MS = 12 * 60 * 60 * 1000

const COUNTRY_NAME_TO_CODE: Record<string, string> = {
  us: 'US',
  usa: 'US',
  'united states': 'US',
  'united states of america': 'US',
  '美国': 'US',
  cn: 'CN',
  china: 'CN',
  '中国': 'CN',
  hk: 'HK',
  'hong kong': 'HK',
  'hong kong sar': 'HK',
  '中国香港': 'HK',
  '香港': 'HK',
  tw: 'TW',
  taiwan: 'TW',
  'taiwan, province of china': 'TW',
  '中国台湾': 'TW',
  '台湾': 'TW',
  mo: 'MO',
  macau: 'MO',
  macao: 'MO',
  '中国澳门': 'MO',
  '澳门': 'MO',
  jp: 'JP',
  japan: 'JP',
  '日本': 'JP',
  kr: 'KR',
  'south korea': 'KR',
  korea: 'KR',
  '韩国': 'KR',
  sg: 'SG',
  singapore: 'SG',
  '新加坡': 'SG',
  my: 'MY',
  malaysia: 'MY',
  '马来西亚': 'MY',
  th: 'TH',
  thailand: 'TH',
  '泰国': 'TH',
  vn: 'VN',
  vietnam: 'VN',
  '越南': 'VN',
  ph: 'PH',
  philippines: 'PH',
  '菲律宾': 'PH',
  id: 'ID',
  indonesia: 'ID',
  '印度尼西亚': 'ID',
  '印尼': 'ID',
  in: 'IN',
  india: 'IN',
  '印度': 'IN',
  gb: 'GB',
  uk: 'GB',
  'united kingdom': 'GB',
  britain: 'GB',
  england: 'GB',
  '英国': 'GB',
  de: 'DE',
  germany: 'DE',
  '德国': 'DE',
  fr: 'FR',
  france: 'FR',
  '法国': 'FR',
  nl: 'NL',
  netherlands: 'NL',
  holland: 'NL',
  '荷兰': 'NL',
  ca: 'CA',
  canada: 'CA',
  '加拿大': 'CA',
  au: 'AU',
  australia: 'AU',
  '澳大利亚': 'AU',
  nz: 'NZ',
  'new zealand': 'NZ',
  '新西兰': 'NZ',
  ru: 'RU',
  russia: 'RU',
  '俄罗斯': 'RU',
  ua: 'UA',
  ukraine: 'UA',
  '乌克兰': 'UA',
  pl: 'PL',
  poland: 'PL',
  '波兰': 'PL',
  cz: 'CZ',
  'czech republic': 'CZ',
  czechia: 'CZ',
  '捷克': 'CZ',
  hu: 'HU',
  hungary: 'HU',
  '匈牙利': 'HU',
  ro: 'RO',
  romania: 'RO',
  '罗马尼亚': 'RO',
  pt: 'PT',
  portugal: 'PT',
  '葡萄牙': 'PT',
  es: 'ES',
  spain: 'ES',
  '西班牙': 'ES',
  it: 'IT',
  italy: 'IT',
  '意大利': 'IT',
  ch: 'CH',
  switzerland: 'CH',
  '瑞士': 'CH',
  se: 'SE',
  sweden: 'SE',
  '瑞典': 'SE',
  no: 'NO',
  norway: 'NO',
  '挪威': 'NO',
  fi: 'FI',
  finland: 'FI',
  '芬兰': 'FI',
  dk: 'DK',
  denmark: 'DK',
  '丹麦': 'DK',
  be: 'BE',
  belgium: 'BE',
  '比利时': 'BE',
  at: 'AT',
  austria: 'AT',
  '奥地利': 'AT',
  ie: 'IE',
  ireland: 'IE',
  '爱尔兰': 'IE',
  br: 'BR',
  brazil: 'BR',
  '巴西': 'BR',
  mx: 'MX',
  mexico: 'MX',
  '墨西哥': 'MX',
  ar: 'AR',
  argentina: 'AR',
  '阿根廷': 'AR',
  tr: 'TR',
  turkey: 'TR',
  '土耳其': 'TR',
  ae: 'AE',
  'united arab emirates': 'AE',
  uae: 'AE',
  '阿联酋': 'AE',
  sa: 'SA',
  'saudi arabia': 'SA',
  '沙特阿拉伯': 'SA',
  qa: 'QA',
  qatar: 'QA',
  '卡塔尔': 'QA',
  il: 'IL',
  israel: 'IL',
  '以色列': 'IL',
  eg: 'EG',
  egypt: 'EG',
  '埃及': 'EG',
  za: 'ZA',
  'south africa': 'ZA',
  '南非': 'ZA',
  cl: 'CL',
  chile: 'CL',
  '智利': 'CL',
}

function normalizeCountryLookupKey(value: string): string {
  return value.trim().toLowerCase().replace(/[().,_-]+/g, ' ').replace(/\s+/g, ' ')
}

function toCountryCodeCandidate(value: unknown): string {
  const text = String(value || '').trim().toUpperCase()
  if (/^[A-Z]{2}$/.test(text)) return text
  return ''
}

export function resolveCountryCode(result?: ProxyIPHealthResult): string {
  if (!result?.ok) return ''

  const rawData = result.rawData || {}
  const directCandidates = [
    rawData.countryCode,
    rawData.country_code,
    rawData.countryCode2,
    rawData.country_code2,
    rawData.country_iso,
    rawData.iso2,
    rawData.iso_2,
    rawData.code,
  ]
  for (const candidate of directCandidates) {
    const code = toCountryCodeCandidate(candidate)
    if (code) return code
  }

  const lookupKeys = [
    result.country,
    rawData.countryName,
    rawData.country_name,
    rawData.country,
  ]

  for (const key of lookupKeys) {
    const normalized = normalizeCountryLookupKey(String(key || ''))
    if (!normalized) continue
    const code = COUNTRY_NAME_TO_CODE[normalized]
    if (code) return code
  }

  return ''
}

export function formatIPLocation(result?: ProxyIPHealthResult): string {
  if (!result?.ok) return ''
  return [result.country, result.region, result.city].filter(Boolean).join(' / ')
}

export function readProxyIPHealthCache(): Record<string, ProxyIPHealthResult> {
  if (typeof window === 'undefined') return {}
  try {
    const raw = localStorage.getItem(PROXY_IP_HEALTH_CACHE_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as { timestamp?: number; data?: Record<string, ProxyIPHealthResult> }
    if (!parsed?.timestamp || !parsed?.data) return {}
    if (Date.now() - parsed.timestamp > PROXY_IP_HEALTH_CACHE_TTL_MS) return {}
    return parsed.data
  } catch {
    return {}
  }
}

export function resolveProxyIPHealthResult(
  proxy: BrowserProxy | undefined,
  cacheMap?: Record<string, ProxyIPHealthResult>,
): ProxyIPHealthResult | undefined {
  if (!proxy) return undefined
  const cacheResult = cacheMap?.[proxy.proxyId]
  if (cacheResult?.ok) return cacheResult
  if (!proxy.lastIPHealthJson) return cacheResult
  try {
    const parsed = JSON.parse(proxy.lastIPHealthJson) as ProxyIPHealthResult
    return parsed
  } catch {
    return cacheResult
  }
}

export function ProxyCountryFlag({
  result,
  className = 'w-4 h-3 rounded-[3px] shadow-sm ring-1 ring-black/5',
}: {
  result?: ProxyIPHealthResult
  className?: string
}) {
  const countryCode = resolveCountryCode(result)
  if (!countryCode) return null

  const FlagComponent = FlagIcons[countryCode as keyof typeof FlagIcons] as ((props: { className?: string; title?: string }) => JSX.Element) | undefined
  if (!FlagComponent) return null

  return <FlagComponent className={className} title={result?.country || countryCode} />
}
