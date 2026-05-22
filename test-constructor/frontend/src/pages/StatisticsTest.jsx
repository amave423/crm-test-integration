import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import "../styles/StatisticsTest.css";
import LogoutButton from "../components/LogoutButton.jsx";
import BackIcon from "../assets/back.svg?react";
import { testsAPI } from "../services/api.js";

export default function StatisticsTest() {
    const { testId } = useParams();
    const navigate = useNavigate();
    const [attempts, setAttempts] = useState([]);
    const [selectedAttempt, setSelectedAttempt] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchAttempts = async () => {
            if (!testId) {
                setLoading(false);
                return;
            }

            try {
                const token = localStorage.getItem("token");
                if (!token) {
                    navigate("/login");
                    return;
                }

                const response = await testsAPI.getTestAttempts(testId);
                const data = response.data;

                if (response.status >= 200 && response.status < 300) {
                    let attemptsArray = [];
                    
                    if (Array.isArray(data)) {
                        attemptsArray = data;
                    } else if (data.attempts && Array.isArray(data.attempts)) {
                        attemptsArray = data.attempts;
                    } else if (data.data && Array.isArray(data.data)) {
                        attemptsArray = data.data;
                    }

                    setAttempts(attemptsArray);
                    console.log("РџРѕР»СѓС‡РµРЅРЅС‹Рµ РїРѕРїС‹С‚РєРё:", attemptsArray);
                } else {
                    console.warn("РќРµ СѓРґР°Р»РѕСЃСЊ Р·Р°РіСЂСѓР·РёС‚СЊ СЃС‚Р°С‚РёСЃС‚РёРєСѓ СЃ СЃРµСЂРІРµСЂР°");
                    setAttempts([]);
                }
            } catch (error) {
                console.error("РћС€РёР±РєР° Р·Р°РіСЂСѓР·РєРё СЃС‚Р°С‚РёСЃС‚РёРєРё:", error);
                setAttempts([]);
            } finally {
                setLoading(false);
            }
        };

        fetchAttempts();
    }, [testId, navigate]);

    const handleBack = () => {
        navigate("/tests");
    };

    const handleOpenDetails = (attempt) => {
        setSelectedAttempt(attempt);
    };

    const handleCloseDetails = () => {
        setSelectedAttempt(null);
    };

    return (
        <div className="tests-page">
            <div
                className="test-page"
                style={{ position: "absolute", left: "1430px", top: "0px" }}
            >
                <LogoutButton />
            </div>
            <div className="create-wrapper2">
                <div className="test2">
                    <div className="stat-top-bar2">
                        <button className="stat-back-btn2" onClick={handleBack}>
                            <BackIcon />
                        </button>
                        <h1>РЎС‚Р°С‚РёСЃС‚РёРєР° С‚РµСЃС‚Р°</h1>

                    </div>
                    <div className="tests-line"></div>
                    {attempts.length === 0 ? (
                        <p className="stat-empty">
                            РџРѕ СЌС‚РѕРјСѓ С‚РµСЃС‚Сѓ РµС‰С‘ РЅРµС‚ РїРѕРїС‹С‚РѕРє.
                        </p>
                    ) : (
                        <div className="stat-attempts-table">
                            <table>
                                <thead>
                                <tr>
                                    <th>РЈС‡Р°СЃС‚РЅРёРє</th>
                                    <th>Р РµР·СѓР»СЊС‚Р°С‚</th>
                                    <th>Р’СЂРµРјСЏ РїСЂРѕС…РѕР¶РґРµРЅРёСЏ</th>
                                    <th>РџРѕРґСЂРѕР±РЅР°СЏ СЃС‚Р°С‚РёСЃС‚РёРєР°</th>
                                </tr>
                                </thead>
                                <tbody>
                                {attempts.map((a) => (
                                    <tr key={a.id}>
                                        <td className="stat-cell-name">{a.userName}</td>
                                        <td className="stat-cell-score">
                                            {a.passed ? "РџСЂРѕР№РґРµРЅ" : "РќРµ РїСЂРѕР№РґРµРЅ"}
                                        </td>

                                        <td className="stat-cell-time">
                                            {a.durationMinutes != null
                                                ? `${a.durationMinutes} РјРёРЅ`
                                                : ""}
                                        </td>
                                        <td className="stat-cell-button">
                                            <button
                                                className="stat-details-btn"
                                                onClick={() => handleOpenDetails(a)}
                                            >
                                                РћС‚РєСЂС‹С‚СЊ РїРѕРґСЂРѕР±РЅСѓСЋ СЃС‚Р°С‚РёСЃС‚РёРєСѓ
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        </div>

                    )}

                    {selectedAttempt && (
                        <div className="stat-modal-overlay">
                            <div className="stat-modal">
                                <h3>РџРѕРґСЂРѕР±РЅР°СЏ СЃС‚Р°С‚РёСЃС‚РёРєР°</h3>

                                <div className="stat-details-user">
                                    <p>
                                        <strong>РЈС‡Р°СЃС‚РЅРёРє:</strong>{" "}
                                        {selectedAttempt.userName}
                                    </p>
                                    <p>
                                        <strong>РџРѕС‡С‚Р°:</strong>{" "}
                                        {selectedAttempt.userEmail}
                                    </p>
                                    <p>
                                        <strong>Р РµР·СѓР»СЊС‚Р°С‚ С‚РµСЃС‚Р°:</strong>{" "}
                                        {selectedAttempt.score}/{selectedAttempt.totalMax}
                                    </p>
                                </div>

                                <div className="stat-table-wrapper">
                                    <table className="stat-table">
                                        <thead>
                                        <tr>
                                            <th>Р’РѕРїСЂРѕСЃ</th>
                                            <th>Р‘Р°Р»Р»С‹</th>
                                        </tr>
                                        </thead>
                                        <tbody>
                                        {(selectedAttempt.perQuestion || []).map((q) => (
                                            <tr key={q.questionIndex}>
                                                <td>{q.questionText}</td>
                                                <td>
                                                    {q.score}/{q.maxScore}
                                                </td>
                                            </tr>
                                        ))}
                                        </tbody>
                                    </table>
                                </div>

                                <button
                                    className="stat-hide-btn"
                                    onClick={handleCloseDetails}
                                >
                                    РЎРєСЂС‹С‚СЊ РїРѕРґСЂРѕР±РЅСѓСЋ СЃС‚Р°С‚РёСЃС‚РёРєСѓ
                                </button>
                            </div>
                        </div>
                    )}

                </div>
            </div>
        </div>
    );
}


