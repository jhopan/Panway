import React, {useEffect, useMemo, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, Grid, Icon, Label, Statistic} from 'semantic-ui-react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import axios from 'axios';
import './Dashboard.css';

const COLORS = {
  requests: '#4318FF',
  quota: '#00B5D8',
  tokens: '#05CD99',
};

const BAR_COLORS = [
  '#4318FF',
  '#00B5D8',
  '#6C63FF',
  '#05CD99',
  '#FFB547',
  '#FF5E7D',
  '#41B883',
  '#7983FF',
  '#FF8F6B',
  '#49BEFF',
];

const DAY_MS = 24 * 60 * 60 * 1000;

const Dashboard = () => {
  const {t, i18n} = useTranslation();
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const response = await axios.get('/api/user/dashboard');
      if (response.data.success) {
        setData(response.data.data || []);
      } else {
        setData([]);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      setData([]);
    } finally {
      setLoading(false);
    }
  };

  // Locale-aware date formatting, follows the active UI language.
  const locale = i18n.language === 'id' ? 'id-ID' : 'en-US';

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString(locale, {month: 'short', day: 'numeric'});
  };

  // Last 7 days, always ending today, with every day present even when empty.
  const {timeSeriesData, modelData, models, summary} = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const days = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date(today.getTime() - i * DAY_MS);
      const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(
        2,
        '0'
      )}-${String(d.getDate()).padStart(2, '0')}`;
      days.push(key);
    }

    const series = {};
    const byModel = {};
    const modelNames = [...new Set(data.map((item) => item.ModelName))];

    days.forEach((day) => {
      series[day] = {date: day, requests: 0, quota: 0, tokens: 0};
      byModel[day] = {date: day};
      modelNames.forEach((m) => {
        byModel[day][m] = 0;
      });
    });

    let totalRequests = 0;
    let totalQuota = 0;
    let totalTokens = 0;

    data.forEach((item) => {
      const day = item.Day;
      if (!series[day]) {
        return;
      }
      const tokens = (item.PromptTokens || 0) + (item.CompletionTokens || 0);
      series[day].requests += item.RequestCount || 0;
      series[day].quota += (item.Quota || 0) / 1000000;
      series[day].tokens += tokens;
      byModel[day][item.ModelName] = tokens;

      totalRequests += item.RequestCount || 0;
      totalQuota += (item.Quota || 0) / 1000000;
      totalTokens += tokens;
    });

    return {
      timeSeriesData: days.map((d) => series[d]),
      modelData: days.map((d) => byModel[d]),
      models: modelNames,
      summary: {totalRequests, totalQuota, totalTokens},
    };
  }, [data]);

  const statCards = [
    {
      key: 'requests',
      icon: 'chart bar',
      color: COLORS.requests,
      label: t('dashboard.summary.requests'),
      value: summary.totalRequests.toLocaleString(locale),
    },
    {
      key: 'quota',
      icon: 'dollar sign',
      color: COLORS.quota,
      label: t('dashboard.summary.quota'),
      value: `$${summary.totalQuota.toFixed(3)}`,
    },
    {
      key: 'tokens',
      icon: 'exchange',
      color: COLORS.tokens,
      label: t('dashboard.summary.tokens'),
      value: summary.totalTokens.toLocaleString(locale),
    },
    {
      key: 'models',
      icon: 'cube',
      color: '#6C63FF',
      label: t('dashboard.summary.models'),
      value: models.length.toLocaleString(locale),
    },
  ];

  const tooltipStyle = {
    background: 'var(--pw-surface, #fff)',
    border: 'none',
    borderRadius: '8px',
    boxShadow: '0 4px 16px rgba(0,0,0,0.12)',
    fontSize: '13px',
  };

  const axisTick = {fontSize: 12, fill: 'var(--pw-faint, #8E9BB3)'};

  // Let recharts drop labels that would collide instead of forcing every
  // day to render: the console is narrower now that the sidebar takes
  // space, so seven day labels overlap on the three small charts.
  const xAxisConfig = {
    dataKey: 'date',
    axisLine: {stroke: 'var(--pw-line, #E4E8F0)'},
    tickLine: false,
    tick: {...axisTick, textAnchor: 'middle'},
    tickFormatter: formatDate,
    interval: 'preserveStartEnd',
    minTickGap: 14,
  };

  const lineChart = (dataKey, color, tooltipKey, formatter) => (
    <ResponsiveContainer width='100%' height={220}>
      <LineChart
        data={timeSeriesData}
        margin={{top: 10, right: 16, left: 4, bottom: 0}}
      >
        <CartesianGrid
          strokeDasharray='3 3'
          vertical={false}
          stroke='var(--pw-line, #EEF1F6)'
        />
        <XAxis {...xAxisConfig} />
        <YAxis
          axisLine={false}
          tickLine={false}
          tick={axisTick}
          width={56}
          tickFormatter={(v) => (formatter ? formatter(v) : v)}
        />
        <Tooltip
          contentStyle={tooltipStyle}
          cursor={{stroke: '#D8DEE9', strokeWidth: 1}}
          formatter={(value) => [
            formatter ? formatter(value) : value,
            t(tooltipKey),
          ]}
          labelFormatter={(label) =>
            `${t('dashboard.statistics.tooltip.date')}: ${formatDate(label)}`
          }
        />
        <Line
          type='monotone'
          dataKey={dataKey}
          stroke={color}
          strokeWidth={2.5}
          dot={{r: 2.5, fill: color, strokeWidth: 0}}
          activeDot={{r: 5, strokeWidth: 2, stroke: 'var(--pw-surface, #fff)'}}
        />
      </LineChart>
    </ResponsiveContainer>
  );

  return (
    <div className='dashboard-container'>
      <Grid columns={4} stackable className='summary-grid'>
        {statCards.map((card) => (
          <Grid.Column key={card.key}>
            <Card fluid className='stat-card'>
              <Card.Content>
                <Statistic>
                  <Statistic.Label>
                    <Icon name={card.icon} />
                    {card.label}
                  </Statistic.Label>
                  <Statistic.Value style={{color: card.color}}>
                    {card.value}
                  </Statistic.Value>
                </Statistic>
              </Card.Content>
            </Card>
          </Grid.Column>
        ))}
      </Grid>

      <Grid columns={3} stackable className='charts-grid'>
        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                <Icon name='chart line' color='blue' />
                {t('dashboard.charts.requests.title')}
              </Card.Header>
              <div className='chart-container'>
                {lineChart(
                  'requests',
                  COLORS.requests,
                  'dashboard.charts.requests.tooltip',
                  (v) => v.toLocaleString(locale)
                )}
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                <Icon name='dollar sign' color='cyan' />
                {t('dashboard.charts.quota.title')}
              </Card.Header>
              <div className='chart-container'>
                {lineChart(
                  'quota',
                  COLORS.quota,
                  'dashboard.charts.quota.tooltip',
                  (v) => `$${Number(v).toFixed(3)}`
                )}
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                <Icon name='exchange' color='green' />
                {t('dashboard.charts.tokens.title')}
              </Card.Header>
              <div className='chart-container'>
                {lineChart(
                  'tokens',
                  COLORS.tokens,
                  'dashboard.charts.tokens.tooltip',
                  (v) => v.toLocaleString(locale)
                )}
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Icon name='cube' color='violet' />
            {t('dashboard.statistics.title')}
            {models.length > 0 && (
              <Label size='small' basic color='blue'>
                {models.length} {t('dashboard.summary.models').toLowerCase()}
              </Label>
            )}
          </Card.Header>
          {loading ? (
            <div className='chart-empty'>{t('dashboard.loading')}</div>
          ) : models.length === 0 ? (
            <div className='chart-empty'>{t('dashboard.no_data')}</div>
          ) : (
            <div className='chart-container'>
              <ResponsiveContainer width='100%' height={320}>
                <BarChart
                  data={modelData}
                  margin={{top: 10, right: 16, left: 0, bottom: 0}}
                >
                  <CartesianGrid
          strokeDasharray='3 3'
          vertical={false}
          stroke='var(--pw-line, #EEF1F6)'
        />
                  <XAxis {...xAxisConfig} />
                  <YAxis
                    axisLine={false}
                    tickLine={false}
                    tick={axisTick}
                    width={56}
                  />
                  <Tooltip
                    contentStyle={tooltipStyle}
                    cursor={{fill: 'rgba(67,24,255,0.04)'}}
                    labelFormatter={(label) =>
                      `${t('dashboard.statistics.tooltip.date')}: ${formatDate(
                        label
                      )}`
                    }
                  />
                  <Legend
                    wrapperStyle={{paddingTop: '16px', fontSize: '13px'}}
                  />
                  {models.map((model, index) => (
                    <Bar
                      key={model}
                      dataKey={model}
                      stackId='a'
                      name={model}
                      fill={BAR_COLORS[index % BAR_COLORS.length]}
                      radius={index === models.length - 1 ? [4, 4, 0, 0] : 0}
                      maxBarSize={46}
                    />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default Dashboard;
