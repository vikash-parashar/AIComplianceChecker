'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { fetchWithAuth } from '../lib/api';

type Violation = {
  issue: string;
  severity: string;
  fix: string;
  rule_type: string;
}

export default function Dashboard() {
  const [tab, setTab] = useState<'sms' | 'policy' | 'config'>('sms');
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(false);
  const [violations, setViolations] = useState<Violation[] | null>(null);
  const [credits, setCredits] = useState<number | null>(null);
  const [billingMsg, setBillingMsg] = useState('');
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/auth');
    } else {
      fetchCredits();
    }
  }, [router]);

  const fetchCredits = async () => {
    try {
      const res = await fetchWithAuth('/billing/credits');
      if (res.ok) {
        const data = await res.json();
        setCredits(data.credits);
      }
    } catch (e) {
      console.error(e);
    }
  };

  const handleBuyCredits = async () => {
    setLoading(true);
    setBillingMsg('');
    try {
      const res = await fetchWithAuth('/billing/checkout', { method: 'POST' });
      const data = await res.json();
      if (res.ok) {
        setBillingMsg(data.message);
        await fetchCredits();
      } else {
        setBillingMsg('Failed to process payment setup.');
      }
    } catch (e) {
      setBillingMsg('Network error on checkout.');
    }
    setLoading(false);
  };


  const handleClear = () => {
    setContent('');
    setViolations(null);
  };

  const handleAnalyze = async () => {
    if (!content.trim()) return;
    
    setLoading(true);
    setViolations(null);

    try {
      let body: any = { content };
      if (tab === 'config') {
        body = { config_json: content };
      }

      const res = await fetchWithAuth(`/analyze/${tab}`, {
        method: 'POST',
        body: JSON.stringify(body)
      });

      if (res.status === 401) {
        localStorage.removeItem('token');
        router.push('/auth');
        return;
      }

      const data = await res.json();
      if (!res.ok) throw new Error(data.error);

      setViolations(data.violations || []);
      await fetchCredits(); // Refresh credits
    } catch (err: any) {
      alert("Failed to analyze: " + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container animate-fade-in">
      <div style={{ textAlign: 'center', marginBottom: '3rem', paddingTop: '2rem' }}>
        <h1 className="title">Compliance Validation Sandbox</h1>
        <p className="subtitle">Identify HIPAA, A2P 10DLC, and GDPR risks before you ship.</p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1.2fr) minmax(0, 1fr)', gap: '2rem' }}>
        {/* Left Side: Input Form */}
        <div className="glass-card" style={{ position: 'relative' }}>
          
          {/* Credits Display overlay */}
          <div style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
              Credits: <strong style={{ color: credits === 0 ? 'var(--danger)' : 'white' }}>{credits !== null ? credits : '...'}</strong>
            </span>
            <button 
              onClick={handleBuyCredits} 
              style={{ background: 'rgba(99, 102, 241, 0.2)', border: '1px solid rgba(99, 102, 241, 0.4)', color: 'white', padding: '0.4rem 0.8rem', borderRadius: '6px', fontSize: '0.8rem', cursor: 'pointer' }}>
              Buy Credits
            </button>
          </div>

          {credits !== null && credits < 3 && credits > 0 && (
            <div style={{ marginBottom: '1rem', padding: '0.8rem', background: 'rgba(245, 158, 11, 0.1)', color: 'var(--warning)', borderRadius: '8px', border: '1px solid rgba(245, 158, 11, 0.2)', fontSize: '0.9rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span>⚠️ <strong>Low Credits:</strong> You only have {credits} scans left.</span>
              <button onClick={handleBuyCredits} style={{ background: 'transparent', border: 'none', color: 'var(--warning)', cursor: 'pointer', textDecoration: 'underline' }}>Top Up Now</button>
            </div>
          )}

          {billingMsg && (
            <div style={{ marginBottom: '1rem', padding: '0.8rem', background: 'rgba(16, 185, 129, 0.1)', color: 'var(--success)', borderRadius: '8px', border: '1px solid rgba(16, 185, 129, 0.2)', fontSize: '0.9rem' }}>
              {billingMsg}
            </div>
          )}

          <div className="tabs" style={{ marginTop: '2.5rem' }}>
            <button className={`tab ${tab === 'sms' ? 'active' : ''}`} onClick={() => setTab('sms')}>
              SMS / Messaging
            </button>
            <button className={`tab ${tab === 'policy' ? 'active' : ''}`} onClick={() => setTab('policy')}>
              Privacy Policy
            </button>
            <button className={`tab ${tab === 'config' ? 'active' : ''}`} onClick={() => setTab('config')}>
              API Config
            </button>
          </div>

          <div className="form-group">
            <label className="form-label" style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span>{tab === 'sms' ? 'Message Content' : tab === 'policy' ? 'Policy Document' : 'JSON Configuration'}</span>
              <span style={{color: 'var(--accent-primary)', textTransform: 'none'}}>{content.length} chars</span>
            </label>
            <textarea 
              className="form-textarea" 
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder={
                tab === 'sms' ? "e.g. 'Hey, your appointment is set for tomorrow.'" :
                tab === 'policy' ? "e.g. 'We store user data indefinitely for diagnostics...'" :
                "e.g. '{\"store_logs\": true, \"phi_retention_days\": null}'"
              }
              style={{ minHeight: '300px' }}
            />
          </div>

          <div style={{ display: 'flex', gap: '1rem' }}>
            <button className="btn-primary" onClick={handleAnalyze} disabled={loading || !content || credits === 0}>
              {loading ? 'Processing...' : credits === 0 ? 'Out of Credits' : 'Scan for Violations'}
            </button>
            {content && (
              <button 
                onClick={handleClear}
                style={{ background: 'transparent', border: '1px solid var(--border)', color: 'var(--text-secondary)', padding: '0.75rem 1.5rem', borderRadius: '8px', cursor: 'pointer', transition: 'all 0.2s' }}
                onMouseEnter={(e) => (e.currentTarget.style.color = 'white')}
                onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-secondary)')}
              >
                Clear
              </button>
            )}
          </div>
        </div>

        {/* Right Side: Results */}
        <div>
          {loading && (
            <div className="glass-card" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '200px' }}>
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
                <div style={{ width: '40px', height: '40px', border: '3px solid var(--border)', borderTopColor: 'var(--accent-primary)', borderRadius: '50%', animation: 'spin 1s linear infinite' }} />
                <p style={{ color: 'var(--text-secondary)' }}>AI Engine Processing Rules...</p>
                <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
              </div>
            </div>
          )}

          {!loading && violations !== null && (
            <div className="animate-fade-in">
              <h3 style={{ fontSize: '1.25rem', marginBottom: '1rem', color: 'white', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                Analysis Results
                {violations.length === 0 ? 
                  <span className="badge badge-low">Passed - 0 Issues</span> : 
                  <span className="badge badge-high">{violations.length} Violations</span>
                }
              </h3>

              {violations.length === 0 ? (
                <div className="glass-card" style={{ borderColor: 'rgba(16, 185, 129, 0.3)', background: 'rgba(16, 185, 129, 0.05)' }}>
                  <p style={{ color: 'var(--success)', fontWeight: '500' }}>✅ No compliance violations found!</p>
                  <p style={{ color: 'var(--text-secondary)', marginTop: '0.5rem', fontSize: '0.9rem' }}>This content appears to adhere to the checked regulations.</p>
                </div>
              ) : (
                <div className="violation-list">
                  {violations.map((v, i) => (
                    <div key={i} className="violation-item">
                      <div className="violation-header">
                        <div>
                          <span className={`badge badge-${v.severity.toLowerCase()}`} style={{ marginRight: '0.75rem' }}>
                            {v.severity.toUpperCase()}
                          </span>
                          <span className="badge" style={{ background: 'rgba(255, 255, 255, 0.1)', color: 'var(--text-secondary)' }}>
                            {v.rule_type.toUpperCase()}
                          </span>
                        </div>
                      </div>
                      <h4 className="violation-issue">{v.issue}</h4>
                      <div className="violation-fix" style={{ marginTop: '1rem' }}>
                        <strong>Suggested Fix</strong>
                        {v.fix}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {!loading && violations === null && (
            <div className="glass-card" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '400px', borderStyle: 'dashed' }}>
              <p style={{ color: 'var(--text-secondary)', textAlign: 'center' }}>
                Enter your text and click <strong>Scan</strong> to see AI analysis.<br/><br/>
                We will test against: <br/>
                <span style={{ color: 'var(--text-primary)' }}>• HIPAA<br/>• A2P 10DLC<br/>• GDPR</span>
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
