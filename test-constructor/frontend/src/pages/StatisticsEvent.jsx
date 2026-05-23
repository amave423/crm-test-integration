import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import '../styles/StatisticsEvent.css';

import LogoutButton from '../components/LogoutButton.jsx';
import BackIcon from '../assets/back.svg?react';
import { eventsAPI } from '../services/api.js';

export default function StatisticsEvent() {
    const navigate = useNavigate();
    const { eventId } = useParams();

    const [selectedStatistic, setSelectedStatistic] = useState(null);
    const [participants, setParticipants] = useState([]);
    const [testHeaders, setTestHeaders] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchStatistics = async () => {
            if (!eventId) {
                setLoading(false);
                setParticipants([]);
                setTestHeaders([]);
                return;
            }

            try {
                setLoading(true);
                setError('');
                const response = await eventsAPI.getEventStatistics(eventId);
                const data = response.data || {};
                setParticipants(Array.isArray(data.participants) ? data.participants : []);
                setTestHeaders(Array.isArray(data.testHeaders) ? data.testHeaders : []);
            } catch (err) {
                console.error('Ошибка загрузки статистики мероприятия:', err);
                setError('Не удалось загрузить статистику мероприятия.');
                setParticipants([]);
                setTestHeaders([]);
            } finally {
                setLoading(false);
            }
        };

        fetchStatistics();
    }, [eventId]);

    const resolvedHeaders = useMemo(() => {
        if (testHeaders.length > 0) return testHeaders;
        const names = new Set();
        participants.forEach(participant => {
            (participant.tests || []).forEach(test => {
                if (test.testName) names.add(test.testName);
            });
        });
        return Array.from(names);
    }, [participants, testHeaders]);

    const handleBack = () => {
        navigate('/events');
    };

    const handleOpenStatistics = (participant, test) => {
        setSelectedStatistic({ participant, test });
    };

    const handleCloseStatistics = () => {
        setSelectedStatistic(null);
    };

    return (
        <div className="tests-page">
            <div className="test-page" style={{ position: 'absolute', left: '1430px', top: '0px' }}>
                <LogoutButton />
            </div>

            <div className="create-wrapper2">
                <div className="test2">
                    <div className="stat-top-bar2">
                        <button className="stat-back-btn2" onClick={handleBack}>
                            <BackIcon />
                        </button>

                        <h1>Статистика мероприятия</h1>
                    </div>

                    <div className="tests-line"></div>

                    {loading ? (
                        <p className="stat-empty">Загрузка статистики...</p>
                    ) : error ? (
                        <p className="stat-empty">{error}</p>
                    ) : participants.length === 0 ? (
                        <p className="stat-empty">По этому мероприятию пока нет результатов тестирования.</p>
                    ) : (
                        <div className="stat-attempts-table">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Участник</th>
                                        {resolvedHeaders.map((test) => (
                                            <th key={test}>{test}</th>
                                        ))}
                                    </tr>
                                </thead>

                                <tbody>
                                    {participants.map((participant) => (
                                        <tr key={participant.id}>
                                            <td className="stat-cell-name">{participant.userName}</td>

                                            {resolvedHeaders.map((header) => {
                                                const test = (participant.tests || []).find(item => item.testName === header);
                                                return (
                                                    <td key={header} className="event-test-cell">
                                                        {test ? (
                                                            <div className="event-test-result">
                                                                <span>{test.score}</span>
                                                                <button
                                                                    className="event-stat-btn"
                                                                    onClick={() => handleOpenStatistics(participant, test)}
                                                                >
                                                                    ▼
                                                                </button>
                                                            </div>
                                                        ) : '—'}
                                                    </td>
                                                );
                                            })}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}

                    {selectedStatistic && (
                        <div className="stat-modal-overlay">
                            <div className="stat-modal">
                                <h3>Подробная статистика</h3>

                                <div className="stat-details-user">
                                    <p><strong>Участник:</strong> {selectedStatistic.participant.userName}</p>
                                    <p><strong>Почта:</strong> {selectedStatistic.participant.email}</p>
                                    <p><strong>Тест:</strong> {selectedStatistic.test.testName}</p>
                                </div>

                                <div className="stat-table-wrapper">
                                    <table className="stat-table">
                                        <thead>
                                            <tr>
                                                <th>Вопрос</th>
                                                <th>Баллы</th>
                                            </tr>
                                        </thead>

                                        <tbody>
                                            {(selectedStatistic.test.questions || []).map((question) => (
                                                <tr key={question.questionIndex}>
                                                    <td>{question.questionText}</td>
                                                    <td>{question.score}/{question.maxScore}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>

                                <button className="stat-hide-btn" onClick={handleCloseStatistics}>
                                    Скрыть подробную статистику
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
