import { useState, useEffect } from "react";
import "../styles/tests.css";
import LogoutButton from "../components/LogoutButton.jsx";
import { useNavigate } from "react-router-dom";

import TaskIcon from "../assets/task.svg?react";
import EventIcon from "../assets/event.svg?react";
import CandidatesIcon from "../assets/Candidates.svg?react";
import SettingsIcon from "../assets/settings.svg?react";
import StatisticsIcon from "../assets/statistics2.svg?react";
import { eventsAPI } from "../services/api";

export default function Events() {
    const navigate = useNavigate();
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchEvents = async () => {
            try {
                setLoading(true);
                setError(null);
                const response = await eventsAPI.getEvents();
                setEvents(Array.isArray(response.data) ? response.data : []);
            } catch (err) {
                console.error("Ошибка загрузки мероприятий:", err);
                setError("Не удалось загрузить мероприятия из CRM.");
                setEvents([]);
            } finally {
                setLoading(false);
            }
        };

        fetchEvents();
    }, []);

    const formatDate = (dateString) => {
        if (!dateString) return "";
        const date = new Date(dateString);
        return Number.isNaN(date.getTime()) ? "" : date.toLocaleDateString("ru-RU");
    };

    return (
        <div className="tests-page">
            <LogoutButton />
            <div className="tests-wrapper">
                <div className="tests">
                    <div className="tests-tabs">
                        <button className="tab-btn" onClick={() => navigate("/tests")}>
                            <TaskIcon />
                            Тестовые задания
                        </button>
                        <button className="tab-btn tab-btn-active" onClick={() => navigate("/events")}>
                            <EventIcon />
                            Мероприятия
                        </button>
                        <button className="tab-btn" onClick={() => navigate("/candidates")}>
                            <CandidatesIcon />
                            Кандидаты
                        </button>
                    </div>

                    <div className="events-container">
                        {loading && <div className="tests-empty">Загрузка мероприятий...</div>}
                        {!loading && error && <div className="tests-empty">{error}</div>}
                        {!loading && !error && events.length === 0 && (
                            <div className="tests-empty">В CRM пока нет мероприятий.</div>
                        )}
                        {!loading && !error && events.map((event) => (
                            <div key={event.id} className="event-card">
                                <div className="event-info">
                                    <h3 className="event-title">{event.name || event.title}</h3>
                                    <div className="event-dates">
                                        <span>Начало: {formatDate(event.start_date || event.startDate)}</span>
                                        <span className="event-separator">|</span>
                                        <span>Конец: {formatDate(event.end_date || event.endDate)}</span>
                                    </div>
                                </div>
                                <div className="event-actions">
                                    <button
                                        className="event-btn settings-btn"
                                        title="Настройка"
                                        onClick={() => navigate(`/event-config?eventId=${event.id}`)}
                                    >
                                        <SettingsIcon />
                                    </button>
                                    <button
                                        className="event-btn statistics-btn"
                                        title="Статистика"
                                        onClick={() => navigate(`/event-statistics/${event.id}`)}
                                    >
                                        <StatisticsIcon />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
