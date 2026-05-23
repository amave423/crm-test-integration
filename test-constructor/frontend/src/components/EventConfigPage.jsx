import React, { useEffect, useMemo, useState } from 'react';
import SelectTestsModal from './details/SelectTestsModal';
import CriteriaTable from './details/CriteriaTable';
import TimeBox from './details/TimeBox';
import ShareLinkBox from './details/ShareLinkBox';
import SpecializationSelect from './details/SpecializationSelect';
import { eventsAPI, testsAPI } from '../services/api.js';
import '../styles/event-config.css';
import back2 from '../assets/back2.svg';
import { useLocation, useNavigate } from 'react-router-dom';
import plusIcon from '../assets/plus.svg';
import korzinaIcon from '../assets/korzina.svg';
import massageIcon from '../assets/message.svg';

const DEFAULT_CRITERIA = [
    { threshold: 75, message: 'Тест пройден', extraTests: [] },
    { threshold: 50, message: 'Назначить дополнительный тест', extraTests: [] },
];

function normalizeTestsPayload(data) {
    let testsArray = [];
    if (Array.isArray(data)) testsArray = data;
    else if (Array.isArray(data?.tests)) testsArray = data.tests;
    else if (Array.isArray(data?.data)) testsArray = data.data;

    return testsArray.map(test => ({
        ...test,
        id: Number(test.test_id || test.id),
        creator_id: test.creator_id ?? test.creatorId ?? test.CreatorID ?? test.creatorID,
        title: test.title || test.name || test.description || `Тест ${test.test_id || test.id}`,
    })).filter(test => Number.isFinite(test.id));
}

function normalizeSpecializationsPayload(data) {
    const items = Array.isArray(data) ? data : data?.specializations;
    return (Array.isArray(items) ? items : [])
        .map(spec => ({
            id: Number(spec.id),
            name: spec.name || spec.title || `Специализация ${spec.id}`,
        }))
        .filter(spec => Number.isFinite(spec.id));
}

function timeToSeconds(time) {
    return Number(time.hours || 0) * 3600 + Number(time.minutes || 0) * 60 + Number(time.seconds || 0);
}

export default function EventConfigPage() {
    const navigate = useNavigate();
    const location = useLocation();
    const eventId = useMemo(() => new URLSearchParams(location.search).get('eventId'), [location.search]);

    const [tests, setTests] = useState([]);
    const [selectedTests, setSelectedTests] = useState([]);
    const [criteria, setCriteria] = useState(DEFAULT_CRITERIA);
    const [modalOpen, setModalOpen] = useState(false);
    const [modalTarget, setModalTarget] = useState(null);
    const [modalSelected, setModalSelected] = useState([]);
    const [specializations, setSpecializations] = useState([]);
    const [selectedSpec, setSelectedSpec] = useState('');
    const [failMessage, setFailMessage] = useState('');
    const [time, setTime] = useState({ hours: 1, minutes: 0, seconds: 0 });
    const [shareLink, setShareLink] = useState('');
    const [saving, setSaving] = useState(false);
    const [statusMessage, setStatusMessage] = useState('');

    useEffect(() => {
        const fetchTests = async () => {
            try {
                const response = await testsAPI.getTests();
                const normalized = normalizeTestsPayload(response.data);
                const userStr = localStorage.getItem('user');
                const currentUserId = userStr ? JSON.parse(userStr).id : null;
                const filtered = currentUserId != null
                    ? normalized.filter(t => Number(t.creator_id) === Number(currentUserId))
                    : normalized;
                setTests(filtered);
            } catch (err) {
                console.error('Ошибка загрузки тестов:', err);
                setTests([]);
            }
        };

        fetchTests();
    }, []);

    useEffect(() => {
        const fetchSpecializations = async () => {
            if (!eventId) {
                setSpecializations([]);
                setSelectedSpec('');
                return;
            }

            try {
                const response = await eventsAPI.getEventSpecializations(eventId);
                const normalized = normalizeSpecializationsPayload(response.data);
                setSpecializations(normalized);
                setSelectedSpec(prev => prev || (normalized[0] ? String(normalized[0].id) : ''));
            } catch (err) {
                console.error('Ошибка загрузки специализаций мероприятия:', err);
                setSpecializations([]);
                setSelectedSpec('');
            }
        };

        fetchSpecializations();
    }, [eventId]);

    const openModal = (target) => {
        setModalTarget(target);
        if (target === 'main') setModalSelected(selectedTests);
        else setModalSelected(criteria[target].extraTests || []);
        setModalOpen(true);
    };

    const handleApplyModal = () => {
        if (modalTarget === 'main') setSelectedTests(modalSelected);
        else {
            setCriteria(criteria.map((row, idx) =>
                idx === modalTarget ? { ...row, extraTests: modalSelected } : row
            ));
        }
        setModalOpen(false);
    };

    const handleCriteriaChange = (idx, newRow) => {
        setCriteria(criteria.map((row, i) => (i === idx ? newRow : row)));
    };

    const handleAddCriteria = () => {
        setCriteria([...criteria, { threshold: 0, message: '', extraTests: [] }]);
    };

    const handleRemoveSelected = (idToRemove) => {
        setSelectedTests(prev => prev.filter(id => id !== idToRemove));
    };

    const handleDeleteCriteria = (index) => {
        setCriteria(prev => prev.filter((_, i) => i !== index));
    };

    const handleDeleteTest = (criteriaIndex, testIndex) => {
        setCriteria(prev => prev.map((item, i) => {
            if (i !== criteriaIndex) return item;
            return { ...item, extraTests: item.extraTests.filter((_, idx) => idx !== testIndex) };
        }));
    };

    const handleSave = async () => {
        setStatusMessage('');
        if (!eventId) {
            setStatusMessage('Не найдено мероприятие CRM.');
            return;
        }
        if (!selectedSpec) {
            setStatusMessage('Выберите специализацию.');
            return;
        }
        if (selectedTests.length === 0) {
            setStatusMessage('Выберите хотя бы один тест.');
            return;
        }

        const mainThreshold = Number(criteria[0]?.threshold || 75);
        const successText = criteria[0]?.message || 'Тест пройден';
        const extraThreshold = criteria.flatMap(row =>
            (row.extraTests || []).map(testId => ({
                threshold: Number(row.threshold || 0),
                message: row.message || '',
                test_id: Number(testId),
            })).filter(item => item.threshold > 0 && item.test_id > 0)
        );

        setSaving(true);
        try {
            const responses = [];
            for (const testId of selectedTests) {
                const response = await eventsAPI.saveEventConfig({
                    event_id: Number(eventId),
                    specialization_id: Number(selectedSpec),
                    test_id: Number(testId),
                    success_text: successText,
                    fail_text: failMessage || 'Тест не пройден',
                    time_limit: timeToSeconds(time),
                    threshold: mainThreshold,
                    extra_threshold: extraThreshold,
                });
                responses.push(response.data);
            }

            const firstLink = responses.find(item => item?.test_link)?.test_link;
            if (firstLink) setShareLink(`${window.location.origin}/test/${firstLink}`);
            setStatusMessage('Настройки тестирования сохранены.');
        } catch (err) {
            console.error('Ошибка сохранения настроек мероприятия:', err);
            setStatusMessage('Не удалось сохранить настройки тестирования.');
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="event-config-page">
            <div className="event-config-sidebar">
                <div className="event-config-header">
                    <button
                        type="button"
                        className="event-config-back-btn"
                        onClick={() => navigate('/events')}
                        aria-label="Вернуться к мероприятиям"
                    >
                        <img src={back2} alt="" className="event-config-back-icon" />
                    </button>
                    <p>Настройка тестов мероприятия</p>
                </div>

                <button className="add-tests-btn" onClick={() => openModal('main')}>
                    <span>Добавить тесты</span>
                    <img src={plusIcon} alt="Добавить" className="add-tests-plus" />
                </button>

                <ul className="event-config-tests-list">
                    {selectedTests.map(id => {
                        const test = tests.find(t => Number(t.id) === Number(id));
                        return test ? (
                            <li key={id} className="event-config-test-item">
                                <span className="test-title">{test.title}</span>
                                <button
                                    className="test-delete-btn"
                                    onClick={() => handleRemoveSelected(id)}
                                    aria-label={`Удалить тест ${test.title}`}
                                    type="button"
                                >
                                    <img src={korzinaIcon} alt="Удалить" />
                                </button>
                            </li>
                        ) : null;
                    })}
                </ul>
            </div>

            <div className="event-config-main">
                <SpecializationSelect
                    specializations={specializations}
                    selected={selectedSpec}
                    onChange={setSelectedSpec}
                />
                <div className="criteria-table-title">Критерии прохождения теста</div>
                <CriteriaTable
                    criteria={criteria}
                    onChange={handleCriteriaChange}
                    onAdd={handleAddCriteria}
                    onAddTest={idx => openModal(idx)}
                    onDelete={handleDeleteCriteria}
                    onDeleteTest={handleDeleteTest}
                    testsList={tests}
                />
                <div className="fail-message-block">
                    <div className="fail-message-header">
                        <img src={massageIcon} alt="" style={{ width: '32px', height: '32px' }} />
                        <p className="fail-message-title">Сообщение при провальном прохождении</p>
                    </div>
                    <input
                        type="text"
                        placeholder="Введите текст сообщения при провальном прохождении..."
                        value={failMessage}
                        onChange={e => setFailMessage(e.target.value)}
                    />
                </div>
                <TimeBox time={time} setTime={setTime} />
                <ShareLinkBox link={shareLink} />
                {statusMessage && <div className="event-config-status">{statusMessage}</div>}
                <button className="save-btn" onClick={handleSave} disabled={saving}>
                    {saving ? 'Сохранение...' : 'Сохранить'}
                </button>
            </div>

            <SelectTestsModal
                open={modalOpen}
                tests={tests}
                selected={modalSelected}
                onSelect={id =>
                    setModalSelected(
                        modalSelected.includes(id)
                            ? modalSelected.filter(i => i !== id)
                            : [...modalSelected, id]
                    )
                }
                onApply={handleApplyModal}
                onClose={() => setModalOpen(false)}
            />
        </div>
    );
}
